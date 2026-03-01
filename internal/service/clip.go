package service

import (
	"errors"

	"renderowl-api/internal/domain"
	"renderowl-api/internal/repository"
)

// ClipService handles clip business logic
type ClipService struct {
	clipRepo     *repository.ClipRepository
	timelineRepo *repository.TimelineRepository
}

// NewClipService creates a new clip service
func NewClipService(clipRepo *repository.ClipRepository, timelineRepo *repository.TimelineRepository) *ClipService {
	return &ClipService{
		clipRepo:     clipRepo,
		timelineRepo: timelineRepo,
	}
}

// Create creates a new clip
func (s *ClipService) Create(userID string, timelineID string, req *CreateClipRequest) (*domain.Clip, error) {
	// Verify timeline belongs to user
	_, err := s.timelineRepo.GetByIDAndUser(timelineID, userID)
	if err != nil {
		return nil, errors.New("timeline not found or access denied")
	}

	clip := &domain.Clip{
		TimelineID:  timelineID,
		TrackID:     req.TrackID,
		Name:        req.Name,
		Type:        req.Type,
		SourceURL:   req.SourceURL,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Duration:    req.EndTime - req.StartTime,
		TrimStart:   req.TrimStart,
		TrimEnd:     req.TrimEnd,
		PositionX:   req.PositionX,
		PositionY:   req.PositionY,
		Scale:       req.Scale,
		Rotation:    req.Rotation,
		Opacity:     req.Opacity,
		TextContent: req.TextContent,
		TextStyle:   req.TextStyle,
	}

	if clip.Scale == 0 {
		clip.Scale = 1
	}
	if clip.Opacity == 0 {
		clip.Opacity = 1
	}

	if err := s.clipRepo.Create(clip); err != nil {
		return nil, err
	}
	return clip, nil
}

// Get retrieves a clip by ID
func (s *ClipService) Get(userID, clipID string) (*domain.Clip, error) {
	clip, err := s.clipRepo.GetByID(clipID)
	if err != nil {
		return nil, err
	}

	// Verify timeline belongs to user
	_, err = s.timelineRepo.GetByIDAndUser(clip.TimelineID, userID)
	if err != nil {
		return nil, errors.New("clip not found or access denied")
	}

	return clip, nil
}

// ListByTimeline lists all clips for a timeline
func (s *ClipService) ListByTimeline(userID, timelineID string) ([]*domain.Clip, error) {
	// Verify timeline belongs to user
	_, err := s.timelineRepo.GetByIDAndUser(timelineID, userID)
	if err != nil {
		return nil, errors.New("timeline not found or access denied")
	}

	return s.clipRepo.ListByTimeline(timelineID)
}

// Update updates a clip
func (s *ClipService) Update(userID, clipID string, req *UpdateClipRequest) (*domain.Clip, error) {
	clip, err := s.Get(userID, clipID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		clip.Name = req.Name
	}
	if req.SourceURL != "" {
		clip.SourceURL = req.SourceURL
	}
	if req.StartTime >= 0 {
		clip.StartTime = req.StartTime
	}
	if req.EndTime > 0 {
		clip.EndTime = req.EndTime
		clip.Duration = req.EndTime - clip.StartTime
	}
	if req.TrimStart >= 0 {
		clip.TrimStart = req.TrimStart
	}
	if req.TrimEnd > 0 {
		clip.TrimEnd = req.TrimEnd
	}
	if req.PositionX != 0 || req.PositionY != 0 {
		clip.PositionX = req.PositionX
		clip.PositionY = req.PositionY
	}
	if req.Scale > 0 {
		clip.Scale = req.Scale
	}
	if req.Rotation != 0 {
		clip.Rotation = req.Rotation
	}
	if req.Opacity > 0 {
		clip.Opacity = req.Opacity
	}
	if req.TextContent != "" {
		clip.TextContent = req.TextContent
	}
	if req.TextStyle != nil {
		clip.TextStyle = req.TextStyle
	}

	if err := s.clipRepo.Update(clip); err != nil {
		return nil, err
	}
	return clip, nil
}

// Delete deletes a clip
func (s *ClipService) Delete(userID, clipID string) error {
	_, err := s.Get(userID, clipID)
	if err != nil {
		return err
	}
	return s.clipRepo.Delete(clipID)
}

// Request types
type CreateClipRequest struct {
	TrackID     string         `json:"trackId" binding:"required"`
	Name        string         `json:"name" binding:"required"`
	Type        string         `json:"type" binding:"required"`
	SourceURL   string         `json:"sourceUrl"`
	StartTime   float64        `json:"startTime" binding:"required"`
	EndTime     float64        `json:"endTime" binding:"required"`
	TrimStart   float64        `json:"trimStart"`
	TrimEnd     float64        `json:"trimEnd"`
	PositionX   float64        `json:"positionX"`
	PositionY   float64        `json:"positionY"`
	Scale       float64        `json:"scale"`
	Rotation    float64        `json:"rotation"`
	Opacity     float64        `json:"opacity"`
	TextContent string         `json:"textContent"`
	TextStyle   *domain.Style  `json:"textStyle"`
}

type UpdateClipRequest struct {
	Name        string        `json:"name"`
	SourceURL   string        `json:"sourceUrl"`
	StartTime   float64       `json:"startTime"`
	EndTime     float64       `json:"endTime"`
	TrimStart   float64       `json:"trimStart"`
	TrimEnd     float64       `json:"trimEnd"`
	PositionX   float64       `json:"positionX"`
	PositionY   float64       `json:"positionY"`
	Scale       float64       `json:"scale"`
	Rotation    float64       `json:"rotation"`
	Opacity     float64       `json:"opacity"`
	TextContent string        `json:"textContent"`
	TextStyle   *domain.Style `json:"textStyle"`
}
