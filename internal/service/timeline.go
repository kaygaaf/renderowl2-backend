package service

import (
	"renderowl-api/internal/domain"
	"renderowl-api/internal/repository"
)

// TimelineService handles timeline business logic
type TimelineService struct {
	repo *repository.TimelineRepository
}

// NewTimelineService creates a new timeline service
func NewTimelineService(repo *repository.TimelineRepository) *TimelineService {
	return &TimelineService{repo: repo}
}

// Create creates a new timeline
func (s *TimelineService) Create(userID string, req *CreateTimelineRequest) (*domain.Timeline, error) {
	timeline := &domain.Timeline{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Duration:    req.Duration,
		Width:       req.Width,
		Height:      req.Height,
		FPS:         req.FPS,
	}

	if timeline.Duration == 0 {
		timeline.Duration = 60
	}
	if timeline.Width == 0 {
		timeline.Width = 1920
	}
	if timeline.Height == 0 {
		timeline.Height = 1080
	}
	if timeline.FPS == 0 {
		timeline.FPS = 30
	}

	if err := s.repo.Create(timeline); err != nil {
		return nil, err
	}
	return timeline, nil
}

// Get retrieves a timeline by ID
func (s *TimelineService) Get(id, userID string) (*domain.Timeline, error) {
	return s.repo.GetByIDAndUser(id, userID)
}

// List retrieves all timelines for a user
func (s *TimelineService) List(userID string, limit, offset int) ([]*domain.Timeline, error) {
	if limit == 0 {
		limit = 20
	}
	return s.repo.ListByUser(userID, limit, offset)
}

// Update updates a timeline
func (s *TimelineService) Update(id, userID string, req *UpdateTimelineRequest) (*domain.Timeline, error) {
	timeline, err := s.repo.GetByIDAndUser(id, userID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		timeline.Name = req.Name
	}
	if req.Description != "" {
		timeline.Description = req.Description
	}
	if req.Duration > 0 {
		timeline.Duration = req.Duration
	}

	if err := s.repo.Update(timeline); err != nil {
		return nil, err
	}
	return timeline, nil
}

// Delete deletes a timeline
func (s *TimelineService) Delete(id, userID string) error {
	_, err := s.repo.GetByIDAndUser(id, userID)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}

// Request types
type CreateTimelineRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Duration    float64 `json:"duration"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	FPS         int     `json:"fps"`
}

type UpdateTimelineRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Duration    float64 `json:"duration"`
}
