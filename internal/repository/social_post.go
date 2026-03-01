package repository

import (
	"context"

	"gorm.io/gorm"
	"renderowl-api/internal/domain/social"
)

// SocialPostRepository implements social.PostRepository
type SocialPostRepository struct {
	db *gorm.DB
}

// NewSocialPostRepository creates a new repository
func NewSocialPostRepository(db *gorm.DB) *SocialPostRepository {
	return &SocialPostRepository{db: db}
}

// Create creates a new scheduled post
func (r *SocialPostRepository) Create(ctx context.Context, post *social.ScheduledPost) error {
	return r.db.WithContext(ctx).Create(post).Error
}

// GetByID gets post by ID
func (r *SocialPostRepository) GetByID(ctx context.Context, id string) (*social.ScheduledPost, error) {
	var post social.ScheduledPost
	err := r.db.WithContext(ctx).First(&post, "id = ?", id).Error
	
	// Load platform posts
	if err == nil {
		var platformPosts []social.PlatformPost
		r.db.WithContext(ctx).Where("scheduled_post_id = ?", id).Find(&platformPosts)
		post.Platforms = platformPosts
	}
	
	return &post, err
}

// GetByUser gets posts by user ID
func (r *SocialPostRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*social.ScheduledPost, error) {
	var posts []*social.ScheduledPost
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("scheduled_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	return posts, err
}

// GetPending gets posts scheduled before a certain time
func (r *SocialPostRepository) GetPending(ctx context.Context, before string) ([]*social.ScheduledPost, error) {
	var posts []*social.ScheduledPost
	err := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", social.PostStatusScheduled, before).
		Find(&posts).Error
	return posts, err
}

// Update updates a post
func (r *SocialPostRepository) Update(ctx context.Context, post *social.ScheduledPost) error {
	return r.db.WithContext(ctx).Save(post).Error
}

// UpdateStatus updates post status
func (r *SocialPostRepository) UpdateStatus(ctx context.Context, id string, status social.PostStatus, errorMsg string) error {
	return r.db.WithContext(ctx).
		Model(&social.ScheduledPost{}).
		Where("id = ?", id).
		Update("status", status).
		Update("error_msg", errorMsg).Error
}

// Delete deletes a post
func (r *SocialPostRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&social.ScheduledPost{}, "id = ?", id).Error
}
