package domain

import (
	"time"
)

// AnalyticsView represents a video view record
type AnalyticsView struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	VideoID   string    `gorm:"index;not null"`
	UserID    string    `gorm:"index"`
	Platform  string    `gorm:"index;not null"` // youtube, tiktok, instagram, etc.
	Count     int64     `gorm:"default:1"`
	Date      time.Time `gorm:"index;not null"`
	IPAddress string
	Country   string
	CreatedAt time.Time
}

// TableName specifies the table name for AnalyticsView
func (AnalyticsView) TableName() string {
	return "analytics_views"
}

// AnalyticsEngagement represents engagement metrics (likes, comments, shares)
type AnalyticsEngagement struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	VideoID   string    `gorm:"index;not null"`
	Platform  string    `gorm:"index;not null"`
	Likes     int64     `gorm:"default:0"`
	Comments  int64     `gorm:"default:0"`
	Shares    int64     `gorm:"default:0"`
	Saves     int64     `gorm:"default:0"`
	Date      time.Time `gorm:"index;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the table name for AnalyticsEngagement
func (AnalyticsEngagement) TableName() string {
	return "analytics_engagement"
}

// UserGrowth tracks user signups and active users
type UserGrowth struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Date        time.Time `gorm:"uniqueIndex;not null"`
	NewSignups  int64     `gorm:"default:0"`
	ActiveUsers int64     `gorm:"default:0"`
	ReturningUsers int64  `gorm:"default:0"`
	ChurnedUsers   int64  `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName specifies the table name for UserGrowth
func (UserGrowth) TableName() string {
	return "analytics_user_growth"
}

// Revenue tracks subscription and credit revenue
type Revenue struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Date         time.Time `gorm:"index;not null"`
	UserID       string    `gorm:"index"`
	Type         string    `gorm:"not null"` // subscription, credits, one_time
	Amount       float64   `gorm:"not null"`
	Currency     string    `gorm:"default:'USD'"`
	Status       string    `gorm:"default:'completed'"` // pending, completed, refunded
	Platform     string    `gorm:"default:'stripe'"`    // stripe, paypal, etc.
	Description  string
	CreatedAt    time.Time
}

// TableName specifies the table name for Revenue
func (Revenue) TableName() string {
	return "analytics_revenue"
}

// VideoPerformance aggregates metrics per video
type VideoPerformance struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	VideoID        string    `gorm:"uniqueIndex;not null"`
	UserID         string    `gorm:"index;not null"`
	Title          string
	TotalViews     int64     `gorm:"default:0"`
	TotalLikes     int64     `gorm:"default:0"`
	TotalComments  int64     `gorm:"default:0"`
	TotalShares    int64     `gorm:"default:0"`
	EngagementRate float64   `gorm:"default:0"`
	Platforms      []string  `gorm:"type:text[]"`
	PublishedAt    *time.Time
	LastUpdated    time.Time
	CreatedAt      time.Time
}

// TableName specifies the table name for VideoPerformance
func (VideoPerformance) TableName() string {
	return "analytics_video_performance"
}

// PlatformStats tracks aggregated stats per platform
type PlatformStats struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Platform     string    `gorm:"uniqueIndex;not null"`
	TotalViews   int64     `gorm:"default:0"`
	TotalVideos  int64     `gorm:"default:0"`
	TotalLikes   int64     `gorm:"default:0"`
	Followers    int64     `gorm:"default:0"`
	LastSyncedAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName specifies the table name for PlatformStats
func (PlatformStats) TableName() string {
	return "analytics_platform_stats"
}

// WebhookEvent stores incoming analytics webhooks
type WebhookEvent struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Platform    string    `gorm:"not null"` // youtube, tiktok, instagram
	EventType   string    `gorm:"not null"` // view, like, share, etc.
	VideoID     string    `gorm:"index"`
	Payload     JSON      `gorm:"type:jsonb"`
	Processed   bool      `gorm:"default:false"`
	ProcessedAt *time.Time
	CreatedAt   time.Time
}

// TableName specifies the table name for WebhookEvent
func (WebhookEvent) TableName() string {
	return "analytics_webhook_events"
}

// JSON is a custom type for JSON fields
type JSON map[string]interface{}
