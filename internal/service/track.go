package service

import (
	"errors"

	"renderowl-api/internal/domain"
	"renderowl-api/internal/repository"
)

// TrackService handles track business logic
type TrackService struct {
	trackRepo    *repository.TrackRepository
	timelineRepo *repository.TimelineRepository
}

// NewTrackService creates a new track service
func NewTrackService(trackRepo *repository.TrackRepository, timelineRepo *repository.TimelineRepository) *TrackService {
	return &TrackService{
		trackRepo:    trackRepo,
		timelineRepo: timelineRepo,
	}
}

// Create creates a new track
func (s *TrackService) Create(userID string, timelineID string, req *CreateTrackRequest) (*domain.Track, error) {
	// Verify timeline belongs to user
	_, err := s.timelineRepo.GetByIDAndUser(timelineID, userID)
	if err != nil {
		return nil, errors.New("timeline not found or access denied")
	}

	// Get current track count for order
	tracks, _ := s.trackRepo.ListByTimeline(timelineID)

	track := &domain.Track{
		TimelineID: timelineID,
		Name:       req.Name,
		Type:       req.Type,
		Order:      len(tracks),
		Muted:      false,
		Solo:       false,
	}

	if err := s.trackRepo.Create(track); err != nil {
		return nil, err
	}
	return track, nil
}

// Get retrieves a track by ID
func (s *TrackService) Get(userID, trackID string) (*domain.Track, error) {
	track, err := s.trackRepo.GetByID(trackID)
	if err != nil {
		return nil, err
	}

	// Verify timeline belongs to user
	_, err = s.timelineRepo.GetByIDAndUser(track.TimelineID, userID)
	if err != nil {
		return nil, errors.New("track not found or access denied")
	}

	return track, nil
}

// ListByTimeline lists all tracks for a timeline
func (s *TrackService) ListByTimeline(userID, timelineID string) ([]*domain.Track, error) {
	// Verify timeline belongs to user
	_, err := s.timelineRepo.GetByIDAndUser(timelineID, userID)
	if err != nil {
		return nil, errors.New("timeline not found or access denied")
	}

	return s.trackRepo.ListByTimeline(timelineID)
}

// Update updates a track
func (s *TrackService) Update(userID, trackID string, req *UpdateTrackRequest) (*domain.Track, error) {
	track, err := s.Get(userID, trackID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		track.Name = req.Name
	}

	if err := s.trackRepo.Update(track); err != nil {
		return nil, err
	}
	return track, nil
}

// Delete deletes a track
func (s *TrackService) Delete(userID, trackID string) error {
	_, err := s.Get(userID, trackID)
	if err != nil {
		return err
	}
	return s.trackRepo.Delete(trackID)
}

// Reorder reorders tracks
func (s *TrackService) Reorder(userID, timelineID string, req *ReorderTracksRequest) error {
	// Verify timeline belongs to user
	_, err := s.timelineRepo.GetByIDAndUser(timelineID, userID)
	if err != nil {
		return errors.New("timeline not found or access denied")
	}

	return s.trackRepo.Reorder(timelineID, req.TrackIDs)
}

// ToggleMute toggles track mute
func (s *TrackService) ToggleMute(userID, trackID string) (*domain.Track, error) {
	_, err := s.Get(userID, trackID)
	if err != nil {
		return nil, err
	}
	return s.trackRepo.ToggleMute(trackID)
}

// ToggleSolo toggles track solo
func (s *TrackService) ToggleSolo(userID, trackID string) (*domain.Track, error) {
	_, err := s.Get(userID, trackID)
	if err != nil {
		return nil, err
	}
	return s.trackRepo.ToggleSolo(trackID)
}

// Request types
type CreateTrackRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required"` // video, audio, text, effect
}

type UpdateTrackRequest struct {
	Name string `json:"name"`
}

type ReorderTracksRequest struct {
	TrackIDs []string `json:"trackIds" binding:"required"`
}
