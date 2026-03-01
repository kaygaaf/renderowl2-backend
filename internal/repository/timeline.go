package repository

import (
	"errors"

	"gorm.io/gorm"

	"renderowl-api/internal/domain"
)

// TimelineRepository defines timeline operations
type TimelineRepository struct {
	db *gorm.DB
}

// NewTimelineRepository creates a new timeline repository
func NewTimelineRepository(db *gorm.DB) *TimelineRepository {
	return &TimelineRepository{db: db}
}

// Create creates a new timeline
func (r *TimelineRepository) Create(timeline *domain.Timeline) error {
	model := toTimelineModel(timeline)
	if err := r.db.Create(model).Error; err != nil {
		return err
	}
	*timeline = *fromTimelineModel(model)
	return nil
}

// GetByID retrieves a timeline by ID
func (r *TimelineRepository) GetByID(id string) (*domain.Timeline, error) {
	var model TimelineModel
	if err := r.db.Preload("Tracks.Clips").First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("timeline not found")
		}
		return nil, err
	}
	return fromTimelineModel(&model), nil
}

// GetByIDAndUser retrieves a timeline by ID and user ID
func (r *TimelineRepository) GetByIDAndUser(id, userID string) (*domain.Timeline, error) {
	var model TimelineModel
	if err := r.db.Preload("Tracks.Clips").Where("id = ? AND user_id = ?", id, userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("timeline not found")
		}
		return nil, err
	}
	return fromTimelineModel(&model), nil
}

// ListByUser lists all timelines for a user
func (r *TimelineRepository) ListByUser(userID string, limit, offset int) ([]*domain.Timeline, error) {
	var models []TimelineModel
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, err
	}

	timelines := make([]*domain.Timeline, len(models))
	for i, m := range models {
		timelines[i] = fromTimelineModel(&m)
	}
	return timelines, nil
}

// Update updates a timeline
func (r *TimelineRepository) Update(timeline *domain.Timeline) error {
	model := toTimelineModel(timeline)
	return r.db.Save(model).Error
}

// Delete deletes a timeline
func (r *TimelineRepository) Delete(id string) error {
	return r.db.Delete(&TimelineModel{}, "id = ?", id).Error
}

// Helper functions
func toTimelineModel(t *domain.Timeline) *TimelineModel {
	return &TimelineModel{
		ID:          t.ID,
		UserID:      t.UserID,
		Name:        t.Name,
		Description: t.Description,
		Duration:    t.Duration,
		Width:       t.Width,
		Height:      t.Height,
		FPS:         t.FPS,
	}
}

func fromTimelineModel(m *TimelineModel) *domain.Timeline {
	t := &domain.Timeline{
		ID:          m.ID,
		UserID:      m.UserID,
		Name:        m.Name,
		Description: m.Description,
		Duration:    m.Duration,
		Width:       m.Width,
		Height:      m.Height,
		FPS:         m.FPS,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	for _, trackModel := range m.Tracks {
		track := domain.Track{
			ID:         trackModel.ID,
			TimelineID: trackModel.TimelineID,
			Name:       trackModel.Name,
			Type:       trackModel.Type,
			Order:      trackModel.Order,
			Muted:      trackModel.Muted,
			Solo:       trackModel.Solo,
		}

		for _, clipModel := range trackModel.Clips {
			clip := domain.Clip{
				ID:          clipModel.ID,
				TimelineID:  clipModel.TimelineID,
				TrackID:     clipModel.TrackID,
				Name:        clipModel.Name,
				Type:        clipModel.Type,
				SourceURL:   clipModel.SourceURL,
				StartTime:   clipModel.StartTime,
				EndTime:     clipModel.EndTime,
				Duration:    clipModel.Duration,
				TrimStart:   clipModel.TrimStart,
				TrimEnd:     clipModel.TrimEnd,
				PositionX:   clipModel.PositionX,
				PositionY:   clipModel.PositionY,
				Scale:       clipModel.Scale,
				Rotation:    clipModel.Rotation,
				Opacity:     clipModel.Opacity,
				TextContent: clipModel.TextContent,
			}
			if clipModel.TextStyle != nil {
				clip.TextStyle = &domain.Style{
					FontSize:   clipModel.TextStyle.FontSize,
					FontFamily: clipModel.TextStyle.FontFamily,
					Color:      clipModel.TextStyle.Color,
					Background: clipModel.TextStyle.Background,
					Bold:       clipModel.TextStyle.Bold,
					Italic:     clipModel.TextStyle.Italic,
					Alignment:  clipModel.TextStyle.Alignment,
				}
			}
			track.Clips = append(track.Clips, clip)
		}

		t.Tracks = append(t.Tracks, track)
	}

	return t
}
