package repository

import (
	"time"
)

// TimelineModel is the database model for timelines
type TimelineModel struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      string `gorm:"index;not null"`
	Name        string `gorm:"not null"`
	Description string
	Duration    float64 `gorm:"default:60"`
	Width       int     `gorm:"default:1920"`
	Height      int     `gorm:"default:1080"`
	FPS         int     `gorm:"default:30"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Tracks      []TrackModel `gorm:"foreignKey:TimelineID;constraint:OnDelete:CASCADE;"`
}

// TableName specifies the table name for TimelineModel
func (TimelineModel) TableName() string {
	return "timelines"
}

// TrackModel is the database model for tracks
type TrackModel struct {
	ID         string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TimelineID string `gorm:"index;not null"`
	Name       string `gorm:"not null"`
	Type       string `gorm:"not null;default:'video'"`
	Order      int    `gorm:"not null;default:0"`
	Muted      bool   `gorm:"default:false"`
	Solo       bool   `gorm:"default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Clips      []ClipModel `gorm:"foreignKey:TrackID;constraint:OnDelete:CASCADE;"`
}

// TableName specifies the table name for TrackModel
func (TrackModel) TableName() string {
	return "tracks"
}

// ClipModel is the database model for clips
type ClipModel struct {
	ID          string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TimelineID  string  `gorm:"index;not null"`
	TrackID     string  `gorm:"index;not null"`
	Name        string  `gorm:"not null"`
	Type        string  `gorm:"not null"`
	SourceURL   string
	StartTime   float64 `gorm:"not null"`
	EndTime     float64 `gorm:"not null"`
	Duration    float64
	TrimStart   float64 `gorm:"default:0"`
	TrimEnd     float64
	PositionX   float64 `gorm:"default:0"`
	PositionY   float64 `gorm:"default:0"`
	Scale       float64 `gorm:"default:1"`
	Rotation    float64 `gorm:"default:0"`
	Opacity     float64 `gorm:"default:1"`
	TextContent string
	TextStyle   *TextStyleModel `gorm:"embedded;embeddedPrefix:text_"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName specifies the table name for ClipModel
func (ClipModel) TableName() string {
	return "clips"
}

// TextStyleModel is the database model for text styling
type TextStyleModel struct {
	FontSize   int
	FontFamily string
	Color      string
	Background string
	Bold       bool
	Italic     bool
	Alignment  string
}
