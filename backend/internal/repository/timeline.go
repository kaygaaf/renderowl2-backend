package repository

import (
	"context"
	"errors"

	"github.com/kaygaaf/renderowl2/internal/domain"
	"gorm.io/gorm"
)

// TimelineRepository defines the interface for timeline data access
type TimelineRepository interface {
	Create(ctx context.Context, timeline *domain.Timeline) error
	GetByID(ctx context.Context, id uint) (*domain.Timeline, error)
	GetByUserID(ctx context.Context, userID uint) ([]domain.Timeline, error)
	Update(ctx context.Context, timeline *domain.Timeline) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]domain.Timeline, error)
}

// timelineRepository implements TimelineRepository
type timelineRepository struct {
	db *gorm.DB
}

// NewTimelineRepository creates a new timeline repository
func NewTimelineRepository(db *gorm.DB) TimelineRepository {
	return &timelineRepository{db: db}
}

// Create creates a new timeline
func (r *timelineRepository) Create(ctx context.Context, timeline *domain.Timeline) error {
	return r.db.WithContext(ctx).Create(timeline).Error
}

// GetByID retrieves a timeline by ID with its associations
func (r *timelineRepository) GetByID(ctx context.Context, id uint) (*domain.Timeline, error) {
	var timeline domain.Timeline
	result := r.db.WithContext(ctx).
		Preload("Tracks.Clips").
		First(&timeline, id)
	
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	
	return &timeline, result.Error
}

// GetByUserID retrieves all timelines for a user
func (r *timelineRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Timeline, error) {
	var timelines []domain.Timeline
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&timelines)
	
	return timelines, result.Error
}

// Update updates a timeline
func (r *timelineRepository) Update(ctx context.Context, timeline *domain.Timeline) error {
	return r.db.WithContext(ctx).Save(timeline).Error
}

// Delete soft-deletes a timeline
func (r *timelineRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Timeline{}, id).Error
}

// List retrieves all timelines with pagination
func (r *timelineRepository) List(ctx context.Context, limit, offset int) ([]domain.Timeline, error) {
	var timelines []domain.Timeline
	result := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&timelines)
	
	return timelines, result.Error
}
