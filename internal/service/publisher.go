package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	socialdomain "renderowl-api/internal/domain/social"
	"renderowl-api/internal/scheduler"
	socialsvc "renderowl-api/internal/service/social"
)

// PostRepository defines post storage operations
type PostRepository interface {
	Create(ctx context.Context, post *socialdomain.ScheduledPost) error
	GetByID(ctx context.Context, id string) (*socialdomain.ScheduledPost, error)
	GetByUser(ctx context.Context, userID string, limit, offset int) ([]*socialdomain.ScheduledPost, error)
	GetPending(ctx context.Context, before string) ([]*socialdomain.ScheduledPost, error)
	Update(ctx context.Context, post *socialdomain.ScheduledPost) error
	UpdateStatus(ctx context.Context, id string, status socialdomain.PostStatus, errorMsg string) error
	Delete(ctx context.Context, id string) error
}

// Publisher handles automatic publishing of scheduled content
type Publisher struct {
	socialService *socialsvc.Service
	scheduler     *scheduler.Scheduler
	postRepo      PostRepository
}

// PublishJobData contains data for a publish job
type PublishJobData struct {
	PostID      string   `json:"postId"`
	AccountID   string   `json:"accountId"`
	VideoPath   string   `json:"videoPath"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Privacy     string   `json:"privacy"`
}

// NewPublisher creates a new publisher instance
func NewPublisher(
	socialService *socialsvc.Service,
	scheduler *scheduler.Scheduler,
	postRepo PostRepository,
) *Publisher {
	return &Publisher{
		socialService: socialService,
		scheduler:     scheduler,
		postRepo:      postRepo,
	}
}

// Initialize sets up the publisher job handlers
func (p *Publisher) Initialize() {
	// Register the publish handler
	p.scheduler.RegisterHandler("publish", p.handlePublishJob)
	p.scheduler.RegisterHandler("crosspost", p.handleCrossPostJob)
}

// SchedulePublish schedules a video for publishing
func (p *Publisher) SchedulePublish(ctx context.Context, post *socialdomain.ScheduledPost) error {
	// Update post status
	post.Status = socialdomain.PostStatusScheduled
	if err := p.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	// Schedule job for each platform
	for _, platformPost := range post.Platforms {
		jobData := PublishJobData{
			PostID:      post.ID,
			AccountID:   platformPost.AccountID,
			VideoPath:   post.Metadata["videoPath"].(string),
			Title:       platformPost.CustomTitle,
			Description: platformPost.CustomDesc,
			Tags:        platformPost.Tags,
			Privacy:     platformPost.Privacy,
		}

		data, _ := json.Marshal(jobData)

		job := &scheduler.Job{
			Name:      "publish",
			Data:      data,
			RunAt:     post.ScheduledAt,
			MaxRetries: 3,
		}

		if err := p.scheduler.AddJob(ctx, job); err != nil {
			return fmt.Errorf("failed to schedule job: %w", err)
		}
	}

	return nil
}

// PublishNow immediately publishes a post
func (p *Publisher) PublishNow(ctx context.Context, postID string) error {
	post, err := p.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("post not found: %w", err)
	}

	// Update status to publishing
	post.Status = socialdomain.PostStatusPublishing
	if err := p.postRepo.Update(ctx, post); err != nil {
		return err
	}

	// Publish to each platform
	for _, platformPost := range post.Platforms {
		go p.publishToPlatform(ctx, post, &platformPost)
	}

	return nil
}

// BulkSchedule schedules multiple posts at once
func (p *Publisher) BulkSchedule(ctx context.Context, posts []*socialdomain.ScheduledPost) error {
	for _, post := range posts {
		if err := p.SchedulePublish(ctx, post); err != nil {
			log.Printf("Failed to schedule post %s: %v", post.ID, err)
			continue
		}
	}
	return nil
}

// GetPublishingQueue returns the current publishing queue
func (p *Publisher) GetPublishingQueue(ctx context.Context, userID string) ([]*socialdomain.ScheduledPost, error) {
	// Get pending posts
	return p.postRepo.GetByUser(ctx, userID, 100, 0)
}

// RetryFailedPost retries a failed post
func (p *Publisher) RetryFailedPost(ctx context.Context, postID string) error {
	post, err := p.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	if post.Status != socialdomain.PostStatusFailed {
		return fmt.Errorf("post is not in failed status")
	}

	// Reset status and reschedule
	post.Status = socialdomain.PostStatusScheduled
	post.ErrorMsg = ""

	if err := p.postRepo.Update(ctx, post); err != nil {
		return err
	}

	return p.SchedulePublish(ctx, post)
}

// Job handlers

func (p *Publisher) handlePublishJob(ctx context.Context, job *scheduler.Job) error {
	var data PublishJobData
	if err := json.Unmarshal(job.Data, &data); err != nil {
		return fmt.Errorf("failed to unmarshal job data: %w", err)
	}

	// Update post status to publishing
	if err := p.postRepo.UpdateStatus(ctx, data.PostID, socialdomain.PostStatusPublishing, ""); err != nil {
		return err
	}

	// Create upload request
	req := &socialdomain.UploadRequest{
		VideoPath:   data.VideoPath,
		Title:       data.Title,
		Description: data.Description,
		Tags:        data.Tags,
		Privacy:     data.Privacy,
	}

	// Upload to platform
	resp, err := p.socialService.UploadVideo(ctx, data.AccountID, req)
	if err != nil {
		// Update post status to failed
		p.postRepo.UpdateStatus(ctx, data.PostID, socialdomain.PostStatusFailed, err.Error())
		return fmt.Errorf("upload failed: %w", err)
	}

	// Update platform post info
	post, _ := p.postRepo.GetByID(ctx, data.PostID)
	for i := range post.Platforms {
		if post.Platforms[i].AccountID == data.AccountID {
			post.Platforms[i].PlatformPostID = resp.PlatformPostID
			post.Platforms[i].PostURL = resp.PostURL
			post.Platforms[i].Status = socialdomain.PostStatusPublished
			post.Platforms[i].PublishedAt = &[]time.Time{time.Now()}[0]
			break
		}
	}

	// Check if all platforms are published
	allPublished := true
	for _, pp := range post.Platforms {
		if pp.Status != socialdomain.PostStatusPublished {
			allPublished = false
			break
		}
	}

	if allPublished {
		post.Status = socialdomain.PostStatusPublished
		post.PublishedAt = &[]time.Time{time.Now()}[0]
	}

	return p.postRepo.Update(ctx, post)
}

func (p *Publisher) handleCrossPostJob(ctx context.Context, job *scheduler.Job) error {
	var data struct {
		PostID      string   `json:"postId"`
		AccountIDs  []string `json:"accountIds"`
		VideoPath   string   `json:"videoPath"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		Privacy     string   `json:"privacy"`
	}

	if err := json.Unmarshal(job.Data, &data); err != nil {
		return err
	}

	req := &socialdomain.UploadRequest{
		VideoPath:   data.VideoPath,
		Title:       data.Title,
		Description: data.Description,
		Tags:        data.Tags,
		Privacy:     data.Privacy,
	}

	// Cross-post to all accounts
	_, err := p.socialService.CrossPost(ctx, data.AccountIDs, req)
	return err
}

// Private methods

func (p *Publisher) publishToPlatform(ctx context.Context, post *socialdomain.ScheduledPost, platformPost *socialdomain.PlatformPost) {
	req := &socialdomain.UploadRequest{
		VideoPath:   post.Metadata["videoPath"].(string),
		Title:       platformPost.CustomTitle,
		Description: platformPost.CustomDesc,
		Tags:        platformPost.Tags,
		Privacy:     platformPost.Privacy,
	}

	resp, err := p.socialService.UploadVideo(ctx, platformPost.AccountID, req)
	if err != nil {
		platformPost.Status = socialdomain.PostStatusFailed
		platformPost.ErrorMsg = err.Error()
		p.postRepo.Update(ctx, post)
		return
	}

	platformPost.PlatformPostID = resp.PlatformPostID
	platformPost.PostURL = resp.PostURL
	platformPost.Status = socialdomain.PostStatusPublished
	now := time.Now()
	platformPost.PublishedAt = &now

	p.postRepo.Update(ctx, post)
}

// FormatForPlatform formats content for a specific platform
func FormatForPlatform(content string, platform socialdomain.SocialPlatform) string {
	switch platform {
	case socialdomain.PlatformTwitter:
		// Twitter has 280 character limit
		if len(content) > 280 {
			return content[:277] + "..."
		}
	case socialdomain.PlatformInstagram:
		// Instagram works well with hashtags
		// Already handled in the platform-specific logic
	case socialdomain.PlatformLinkedIn:
		// LinkedIn prefers professional tone
		// Content can be longer
	case socialdomain.PlatformTikTok:
		// TikTok description limit is 2200 characters
		if len(content) > 2200 {
			return content[:2197] + "..."
		}
	}
	return content
}
