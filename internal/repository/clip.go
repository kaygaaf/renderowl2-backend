package repository

import (
	"errors"

	"gorm.io/gorm"

	"renderowl-api/internal/domain"
)

// ClipRepository defines clip operations
type ClipRepository struct {
	db *gorm.DB
}

// NewClipRepository creates a new clip repository
func NewClipRepository(db *gorm.DB) *ClipRepository {
	return &ClipRepository{db: db}
}

// Create creates a new clip
func (r *ClipRepository) Create(clip *domain.Clip) error {
	model := toClipModel(clip)
	if err := r.db.Create(model).Error; err != nil {
		return err
	}
	*clip = *fromClipModel(model)
	return nil
}

// GetByID retrieves a clip by ID
func (r *ClipRepository) GetByID(id string) (*domain.Clip, error) {
	var model ClipModel
	if err := r.db.First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("clip not found")
		}
		return nil, err
	}
	return fromClipModel(&model), nil
}

// ListByTimeline lists all clips for a timeline
func (r *ClipRepository) ListByTimeline(timelineID string) ([]*domain.Clip, error) {
	var models []ClipModel
	if err := r.db.Where("timeline_id = ?", timelineID).Order("start_time ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	clips := make([]*domain.Clip, len(models))
	for i, m := range models {
		clips[i] = fromClipModel(&m)
	}
	return clips, nil
}

// ListByTrack lists all clips for a track
func (r *ClipRepository) ListByTrack(trackID string) ([]*domain.Clip, error) {
	var models []ClipModel
	if err := r.db.Where("track_id = ?", trackID).Order("start_time ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	clips := make([]*domain.Clip, len(models))
	for i, m := range models {
		clips[i] = fromClipModel(&m)
	}
	return clips, nil
}

// Update updates a clip
func (r *ClipRepository) Update(clip *domain.Clip) error {
	model := toClipModel(clip)
	return r.db.Save(model).Error
}

// Delete deletes a clip
func (r *ClipRepository) Delete(id string) error {
	return r.db.Delete(&ClipModel{}, "id = ?", id).Error
}

// Helper functions
func toClipModel(c *domain.Clip) *ClipModel {
	m := &ClipModel{
		ID:          c.ID,
		TimelineID:  c.TimelineID,
		TrackID:     c.TrackID,
		Name:        c.Name,
		Type:        c.Type,
		SourceURL:   c.SourceURL,
		StartTime:   c.StartTime,
		EndTime:     c.EndTime,
		Duration:    c.Duration,
		TrimStart:   c.TrimStart,
		TrimEnd:     c.TrimEnd,
		PositionX:   c.PositionX,
		PositionY:   c.PositionY,
		Scale:       c.Scale,
		Rotation:    c.Rotation,
		Opacity:     c.Opacity,
		TextContent: c.TextContent,
	}
	if c.TextStyle != nil {
		m.TextStyle = &TextStyleModel{
			FontSize:   c.TextStyle.FontSize,
			FontFamily: c.TextStyle.FontFamily,
			Color:      c.TextStyle.Color,
			Background: c.TextStyle.Background,
			Bold:       c.TextStyle.Bold,
			Italic:     c.TextStyle.Italic,
			Alignment:  c.TextStyle.Alignment,
		}
	}
	return m
}

func fromClipModel(m *ClipModel) *domain.Clip {
	c := &domain.Clip{
		ID:          m.ID,
		TimelineID:  m.TimelineID,
		TrackID:     m.TrackID,
		Name:        m.Name,
		Type:        m.Type,
		SourceURL:   m.SourceURL,
		StartTime:   m.StartTime,
		EndTime:     m.EndTime,
		Duration:    m.Duration,
		TrimStart:   m.TrimStart,
		TrimEnd:     m.TrimEnd,
		PositionX:   m.PositionX,
		PositionY:   m.PositionY,
		Scale:       m.Scale,
		Rotation:    m.Rotation,
		Opacity:     m.Opacity,
		TextContent: m.TextContent,
	}
	if m.TextStyle != nil {
		c.TextStyle = &domain.Style{
			FontSize:   m.TextStyle.FontSize,
			FontFamily: m.TextStyle.FontFamily,
			Color:      m.TextStyle.Color,
			Background: m.TextStyle.Background,
			Bold:       m.TextStyle.Bold,
			Italic:     m.TextStyle.Italic,
			Alignment:  m.TextStyle.Alignment,
		}
	}
	return c
}
