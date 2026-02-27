package domain

import (
	"time"

	"gorm.io/gorm"
)

// Timeline represents a user's timeline/project
type Timeline struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description"`
	Status      TimelineStatus `json:"status" gorm:"default:active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Associations
	Tracks []Track `json:"tracks,omitempty" gorm:"foreignKey:TimelineID"`
}

// TimelineStatus represents the status of a timeline
type TimelineStatus string

const (
	TimelineStatusActive   TimelineStatus = "active"
	TimelineStatusArchived TimelineStatus = "archived"
	TimelineStatusDeleted  TimelineStatus = "deleted"
)

// Track represents a media track in a timeline
type Track struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	TimelineID uint           `json:"timeline_id" gorm:"not null;index"`
	Name       string         `json:"name" gorm:"not null"`
	Type       TrackType      `json:"type" gorm:"not null"`
	Order      int            `json:"order" gorm:"default:0"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Associations
	Clips []Clip `json:"clips,omitempty" gorm:"foreignKey:TrackID"`
}

// TrackType represents the type of track
type TrackType string

const (
	TrackTypeVideo TrackType = "video"
	TrackTypeAudio TrackType = "audio"
	TrackTypeText  TrackType = "text"
)

// Clip represents a media clip in a track
type Clip struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	TrackID     uint           `json:"track_id" gorm:"not null;index"`
	Name        string         `json:"name"`
	SourceURL   string         `json:"source_url"`
	StartTime   float64        `json:"start_time" gorm:"default:0"`
	EndTime     float64        `json:"end_time"`
	Duration    float64        `json:"duration"`
	Position    float64        `json:"position" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// CreateTimelineRequest represents a request to create a timeline
type CreateTimelineRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// UpdateTimelineRequest represents a request to update a timeline
type UpdateTimelineRequest struct {
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	Status      TimelineStatus `json:"status,omitempty"`
}

// TimelineResponse represents a timeline response
type TimelineResponse struct {
	ID          uint           `json:"id"`
	UserID      uint           `json:"user_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      TimelineStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	TrackCount  int            `json:"track_count,omitempty"`
}
