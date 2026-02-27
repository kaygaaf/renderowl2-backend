package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/kaygaaf/renderowl2/internal/domain"
	"github.com/kaygaaf/renderowl2/internal/repository"
)

// TimelineService defines the interface for timeline business logic
type TimelineService interface {
	Create(ctx context.Context, userID uint, req domain.CreateTimelineRequest) (*domain.Timeline, error)
	GetByID(ctx context.Context, id uint) (*domain.Timeline, error)
	GetByUserID(ctx context.Context, userID uint) ([]domain.Timeline, error)
	Update(ctx context.Context, id uint, userID uint, req domain.UpdateTimelineRequest) (*domain.Timeline, error)
	Delete(ctx context.Context, id uint, userID uint) error
	List(ctx context.Context, limit, offset int) ([]domain.Timeline, error)
}

// timelineService implements TimelineService
type timelineService struct {
	repo repository.TimelineRepository
}

// NewTimelineService creates a new timeline service
func NewTimelineService(repo repository.TimelineRepository) TimelineService {
	return &timelineService{repo: repo}
}

// Create creates a new timeline
func (s *timelineService) Create(ctx context.Context, userID uint, req domain.CreateTimelineRequest) (*domain.Timeline, error) {
	timeline := &domain.Timeline{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      domain.TimelineStatusActive,
	}

	if err := s.repo.Create(ctx, timeline); err != nil {
		return nil, fmt.Errorf("failed to create timeline: %w", err)
	}

	return timeline, nil
}

// GetByID retrieves a timeline by ID
func (s *timelineService) GetByID(ctx context.Context, id uint) (*domain.Timeline, error) {
	timeline, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline: %w", err)
	}

	if timeline == nil {
		return nil, errors.New("timeline not found")
	}

	return timeline, nil
}

// GetByUserID retrieves all timelines for a user
func (s *timelineService) GetByUserID(ctx context.Context, userID uint) ([]domain.Timeline, error) {
	timelines, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user timelines: %w", err)
	}

	return timelines, nil
}

// Update updates a timeline
func (s *timelineService) Update(ctx context.Context, id uint, userID uint, req domain.UpdateTimelineRequest) (*domain.Timeline, error) {
	timeline, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline: %w", err)
	}

	if timeline == nil {
		return nil, errors.New("timeline not found")
	}

	// Check ownership
	if timeline.UserID != userID {
		return nil, errors.New("unauthorized: not your timeline")
	}

	// Update fields if provided
	if req.Title != "" {
		timeline.Title = req.Title
	}
	if req.Description != "" {
		timeline.Description = req.Description
	}
	if req.Status != "" {
		timeline.Status = req.Status
	}

	if err := s.repo.Update(ctx, timeline); err != nil {
		return nil, fmt.Errorf("failed to update timeline: %w", err)
	}

	return timeline, nil
}

// Delete deletes a timeline
func (s *timelineService) Delete(ctx context.Context, id uint, userID uint) error {
	timeline, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get timeline: %w", err)
	}

	if timeline == nil {
		return errors.New("timeline not found")
	}

	// Check ownership
	if timeline.UserID != userID {
		return errors.New("unauthorized: not your timeline")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete timeline: %w", err)
	}

	return nil
}

// List retrieves all timelines with pagination
func (s *timelineService) List(ctx context.Context, limit, offset int) ([]domain.Timeline, error) {
	timelines, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list timelines: %w", err)
	}

	return timelines, nil
}

// ToResponse converts a Timeline to TimelineResponse
func ToResponse(t *domain.Timeline) domain.TimelineResponse {
	return domain.TimelineResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		TrackCount:  len(t.Tracks),
	}
}

// ToResponseList converts a list of Timelines to TimelineResponses
func ToResponseList(timelines []domain.Timeline) []domain.TimelineResponse {
	responses := make([]domain.TimelineResponse, len(timelines))
	for i, t := range timelines {
		responses[i] = ToResponse(&t)
	}
	return responses
}
