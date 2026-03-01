package domain

import (
	"time"
)

// Timeline represents a video timeline
type Timeline struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Duration    float64   `json:"duration"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	FPS         int       `json:"fps"`
	Tracks      []Track   `json:"tracks,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Track represents a track in a timeline
type Track struct {
	ID         string  `json:"id"`
	TimelineID string  `json:"timelineId"`
	Name       string  `json:"name"`
	Type       string  `json:"type"` // video, audio, text, effect
	Order      int     `json:"order"`
	Muted      bool    `json:"muted"`
	Solo       bool    `json:"solo"`
	Clips      []Clip  `json:"clips,omitempty"`
}

// Clip represents a media clip on a track
type Clip struct {
	ID           string  `json:"id"`
	TimelineID   string  `json:"timelineId"`
	TrackID      string  `json:"trackId"`
	Name         string  `json:"name"`
	Type         string  `json:"type"` // video, audio, image, text
	SourceURL    string  `json:"sourceUrl"`
	StartTime    float64 `json:"startTime"`
	EndTime      float64 `json:"endTime"`
	Duration     float64 `json:"duration"`
	TrimStart    float64 `json:"trimStart"`
	TrimEnd      float64 `json:"trimEnd"`
	PositionX    float64 `json:"positionX"`
	PositionY    float64 `json:"positionY"`
	Scale        float64 `json:"scale"`
	Rotation     float64 `json:"rotation"`
	Opacity      float64 `json:"opacity"`
	TextContent  string  `json:"textContent,omitempty"`
	TextStyle    *Style  `json:"textStyle,omitempty"`
}

// Style represents styling for text clips
type Style struct {
	FontSize   int     `json:"fontSize"`
	FontFamily string  `json:"fontFamily"`
	Color      string  `json:"color"`
	Background string  `json:"background"`
	Bold       bool    `json:"bold"`
	Italic     bool    `json:"italic"`
	Alignment  string  `json:"alignment"`
}

// UserContext holds authenticated user info
type UserContext struct {
	ID    string
	Email string
}
