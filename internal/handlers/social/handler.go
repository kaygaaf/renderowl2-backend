package social

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	socialdomain "renderowl-api/internal/domain/social"
	"renderowl-api/internal/scheduler"
	"renderowl-api/internal/service"
	socialsvc "renderowl-api/internal/service/social"
)

// Handler handles social media HTTP requests
type Handler struct {
	socialService *socialsvc.Service
	publisher     *service.Publisher
	scheduler     *scheduler.Scheduler
}

// NewSocialHandler creates a new social media handler
func NewSocialHandler(
	socialService *socialsvc.Service,
	publisher *service.Publisher,
	scheduler *scheduler.Scheduler,
) *Handler {
	return &Handler{
		socialService: socialService,
		publisher:     publisher,
		scheduler:     scheduler,
	}
}

// GetPlatforms returns available platforms
func (h *Handler) GetPlatforms(c *gin.Context) {
	platforms := h.socialService.GetPlatforms()
	c.JSON(http.StatusOK, gin.H{
		"platforms": platforms,
	})
}

// GetAccounts returns connected accounts for the user
func (h *Handler) GetAccounts(c *gin.Context) {
	userID := c.GetString("userID")

	accounts, err := h.socialService.GetAccounts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}

// GetAccount returns a specific account
func (h *Handler) GetAccount(c *gin.Context) {
	accountID := c.Param("id")

	account, err := h.socialService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// DisconnectAccount removes a connected account
func (h *Handler) DisconnectAccount(c *gin.Context) {
	accountID := c.Param("id")

	if err := h.socialService.DisconnectAccount(c.Request.Context(), accountID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account disconnected"})
}

// GetAuthURL returns OAuth URL for a platform
func (h *Handler) GetAuthURL(c *gin.Context) {
	platform := socialdomain.SocialPlatform(c.Param("platform"))
	state := c.Query("state")
	if state == "" {
		state = generateState()
	}

	authURL, err := h.socialService.GetAuthURL(platform, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":   authURL,
		"state": state,
	})
}

// HandleCallback handles OAuth callback
func (h *Handler) HandleCallback(c *gin.Context) {
	platform := socialdomain.SocialPlatform(c.Param("platform"))
	userID := c.GetString("userID")

	var req struct {
		Code string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	account, err := h.socialService.ConnectAccount(c.Request.Context(), platform, req.Code, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

// UploadVideo uploads a video immediately
func (h *Handler) UploadVideo(c *gin.Context) {
	var req struct {
		AccountID   string   `json:"accountId"`
		VideoPath   string   `json:"videoPath"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		Privacy     string   `json:"privacy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	uploadReq := &socialdomain.UploadRequest{
		VideoPath:   req.VideoPath,
		Title:       req.Title,
		Description: req.Description,
		Tags:        req.Tags,
		Privacy:     req.Privacy,
	}

	resp, err := h.socialService.UploadVideo(c.Request.Context(), req.AccountID, uploadReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CrossPost uploads to multiple platforms
func (h *Handler) CrossPost(c *gin.Context) {
	var req struct {
		AccountIDs  []string `json:"accountIds"`
		VideoPath   string   `json:"videoPath"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		Privacy     string   `json:"privacy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	uploadReq := &socialdomain.UploadRequest{
		VideoPath:   req.VideoPath,
		Title:       req.Title,
		Description: req.Description,
		Tags:        req.Tags,
		Privacy:     req.Privacy,
	}

	results, err := h.socialService.CrossPost(c.Request.Context(), req.AccountIDs, uploadReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// SchedulePost schedules a post for later
func (h *Handler) SchedulePost(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		VideoID     string                   `json:"videoId"`
		Title       string                   `json:"title"`
		Description string                   `json:"description"`
		Platforms   []PlatformScheduleReq    `json:"platforms"`
		ScheduledAt string                   `json:"scheduledAt"`
		Timezone    string                   `json:"timezone"`
		Recurring   *socialdomain.RecurringRule `json:"recurring,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	scheduledAt, err := parseTime(req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled time"})
		return
	}

	post := &socialdomain.ScheduledPost{
		UserID:      userID,
		VideoID:     req.VideoID,
		Title:       req.Title,
		Description: req.Description,
		ScheduledAt: scheduledAt,
		Timezone:    req.Timezone,
		Recurring:   req.Recurring,
		Metadata: socialdomain.JSON{
			"videoPath": req.VideoID, // Would be resolved from video service
		},
	}

	// Convert platform requests
	for _, p := range req.Platforms {
		post.Platforms = append(post.Platforms, socialdomain.PlatformPost{
			AccountID:   p.AccountID,
			Platform:    socialdomain.SocialPlatform(p.Platform),
			CustomTitle: p.Title,
			CustomDesc:  p.Description,
			Tags:        p.Tags,
			Privacy:     p.Privacy,
		})
	}

	if err := h.socialService.SchedulePost(c.Request.Context(), post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Schedule with publisher
	if err := h.publisher.SchedulePublish(c.Request.Context(), post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// GetScheduledPosts returns scheduled posts
func (h *Handler) GetScheduledPosts(c *gin.Context) {
	userID := c.GetString("userID")

	posts, err := h.socialService.GetScheduledPosts(c.Request.Context(), userID, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

// CancelScheduledPost cancels a scheduled post
func (h *Handler) CancelScheduledPost(c *gin.Context) {
	userID := c.GetString("userID")
	postID := c.Param("id")

	if err := h.socialService.CancelScheduledPost(c.Request.Context(), postID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post cancelled"})
}

// PublishNow publishes a scheduled post immediately
func (h *Handler) PublishNow(c *gin.Context) {
	postID := c.Param("id")

	if err := h.publisher.PublishNow(c.Request.Context(), postID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Publishing started"})
}

// RetryPost retries a failed post
func (h *Handler) RetryPost(c *gin.Context) {
	postID := c.Param("id")

	if err := h.publisher.RetryFailedPost(c.Request.Context(), postID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post queued for retry"})
}

// GetPublishingQueue returns the publishing queue
func (h *Handler) GetPublishingQueue(c *gin.Context) {
	userID := c.GetString("userID")

	posts, err := h.publisher.GetPublishingQueue(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"queue": posts,
	})
}

// GetAnalytics returns analytics for an account
func (h *Handler) GetAnalytics(c *gin.Context) {
	accountID := c.Param("accountId")
	postID := c.Query("postId")

	analytics, err := h.socialService.GetAnalytics(c.Request.Context(), accountID, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetTrends returns trends for a platform
func (h *Handler) GetTrends(c *gin.Context) {
	accountID := c.Param("accountId")
	region := c.DefaultQuery("region", "US")

	trends, err := h.socialService.GetTrends(c.Request.Context(), accountID, region)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
	})
}

// GetQueueStats returns queue statistics
func (h *Handler) GetQueueStats(c *gin.Context) {
	stats, err := h.scheduler.GetQueueStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Helper types and functions

type PlatformScheduleReq struct {
	AccountID   string   `json:"accountId"`
	Platform    string   `json:"platform"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Privacy     string   `json:"privacy"`
}

func generateState() string {
	// Generate random state string
	return "state_" + generateID()
}

func generateID() string {
	// Simple ID generation
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func parseTime(s string) (time.Time, error) {
	// Parse ISO 8601 time
	return time.Parse(time.RFC3339, s)
}
