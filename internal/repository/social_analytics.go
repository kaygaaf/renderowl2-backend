package repository

import (
	"context"

	"gorm.io/gorm"
	"renderowl-api/internal/domain/social"
)

// SocialAnalyticsRepository implements social.AnalyticsRepository
type SocialAnalyticsRepository struct {
	db *gorm.DB
}

// NewSocialAnalyticsRepository creates a new repository
func NewSocialAnalyticsRepository(db *gorm.DB) *SocialAnalyticsRepository {
	return &SocialAnalyticsRepository{db: db}
}

// Create creates analytics data
func (r *SocialAnalyticsRepository) Create(ctx context.Context, data *social.AnalyticsData) error {
	return r.db.WithContext(ctx).Create(data).Error
}

// GetByPost gets analytics by post ID
func (r *SocialAnalyticsRepository) GetByPost(ctx context.Context, postID string) ([]*social.AnalyticsData, error) {
	var data []*social.AnalyticsData
	err := r.db.WithContext(ctx).
		Where("post_id = ?", postID).
		Order("recorded_at DESC").
		Find(&data).Error
	return data, err
}

// GetLatestByPost gets latest analytics for a post
func (r *SocialAnalyticsRepository) GetLatestByPost(ctx context.Context, postID string) (*social.AnalyticsData, error) {
	var data social.AnalyticsData
	err := r.db.WithContext(ctx).
		Where("post_id = ?", postID).
		Order("recorded_at DESC").
		First(&data).Error
	return &data, err
}
