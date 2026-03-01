package repository

import (
	"errors"

	"gorm.io/gorm"

	"renderowl-api/internal/domain"
)

// TrackRepository defines track operations
type TrackRepository struct {
	db *gorm.DB
}

// NewTrackRepository creates a new track repository
func NewTrackRepository(db *gorm.DB) *TrackRepository {
	return &TrackRepository{db: db}
}

// Create creates a new track
func (r *TrackRepository) Create(track *domain.Track) error {
	model := toTrackModel(track)
	if err := r.db.Create(model).Error; err != nil {
		return err
	}
	*track = *fromTrackModel(model)
	return nil
}

// GetByID retrieves a track by ID
func (r *TrackRepository) GetByID(id string) (*domain.Track, error) {
	var model TrackModel
	if err := r.db.First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("track not found")
		}
		return nil, err
	}
	return fromTrackModel(&model), nil
}

// ListByTimeline lists all tracks for a timeline
func (r *TrackRepository) ListByTimeline(timelineID string) ([]*domain.Track, error) {
	var models []TrackModel
	if err := r.db.Where("timeline_id = ?", timelineID).Order("\"order\" ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	tracks := make([]*domain.Track, len(models))
	for i, m := range models {
		tracks[i] = fromTrackModel(&m)
	}
	return tracks, nil
}

// Update updates a track
func (r *TrackRepository) Update(track *domain.Track) error {
	model := toTrackModel(track)
	return r.db.Save(model).Error
}

// Delete deletes a track
func (r *TrackRepository) Delete(id string) error {
	return r.db.Delete(&TrackModel{}, "id = ?", id).Error
}

// Reorder reorders tracks
func (r *TrackRepository) Reorder(timelineID string, trackIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, trackID := range trackIDs {
			if err := tx.Model(&TrackModel{}).Where("id = ? AND timeline_id = ?", trackID, timelineID).Update("\"order\"", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ToggleMute toggles track mute status
func (r *TrackRepository) ToggleMute(id string) (*domain.Track, error) {
	var model TrackModel
	if err := r.db.First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	model.Muted = !model.Muted
	if err := r.db.Save(&model).Error; err != nil {
		return nil, err
	}
	return fromTrackModel(&model), nil
}

// ToggleSolo toggles track solo status
func (r *TrackRepository) ToggleSolo(id string) (*domain.Track, error) {
	var model TrackModel
	if err := r.db.First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	model.Solo = !model.Solo
	if err := r.db.Save(&model).Error; err != nil {
		return nil, err
	}
	return fromTrackModel(&model), nil
}

// Helper functions
func toTrackModel(t *domain.Track) *TrackModel {
	return &TrackModel{
		ID:         t.ID,
		TimelineID: t.TimelineID,
		Name:       t.Name,
		Type:       t.Type,
		Order:      t.Order,
		Muted:      t.Muted,
		Solo:       t.Solo,
	}
}

func fromTrackModel(m *TrackModel) *domain.Track {
	return &domain.Track{
		ID:         m.ID,
		TimelineID: m.TimelineID,
		Name:       m.Name,
		Type:       m.Type,
		Order:      m.Order,
		Muted:      m.Muted,
		Solo:       m.Solo,
	}
}
