package social

import (
	"time"
)

// SocialPlatform represents a supported social media platform
type SocialPlatform string

const (
	PlatformYouTube   SocialPlatform = "youtube"
	PlatformTikTok    SocialPlatform = "tiktok"
	PlatformInstagram SocialPlatform = "instagram"
	PlatformTwitter   SocialPlatform = "twitter"
	PlatformLinkedIn  SocialPlatform = "linkedin"
	PlatformFacebook  SocialPlatform = "facebook"
)

// PlatformStatus represents the connection status of a platform
type PlatformStatus string

const (
	StatusConnected    PlatformStatus = "connected"
	StatusDisconnected PlatformStatus = "disconnected"
	StatusExpired      PlatformStatus = "expired"
	StatusError        PlatformStatus = "error"
)

// PostStatus represents the status of a scheduled post
type PostStatus string

const (
	PostStatusDraft      PostStatus = "draft"
	PostStatusScheduled  PostStatus = "scheduled"
	PostStatusPublishing PostStatus = "publishing"
	PostStatusPublished  PostStatus = "published"
	PostStatusFailed     PostStatus = "failed"
	PostStatusCancelled  PostStatus = "cancelled"
)

// SocialAccount represents a connected social media account
type SocialAccount struct {
	ID           string         `json:"id" gorm:"primaryKey"`
	UserID       string         `json:"userId" gorm:"index"`
	Platform     SocialPlatform `json:"platform"`
	AccountID    string         `json:"accountId"`
	AccountName  string         `json:"accountName"`
	AccessToken  string         `json:"-" gorm:"column:access_token"`
	RefreshToken string         `json:"-" gorm:"column:refresh_token"`
	TokenExpiry  *time.Time     `json:"tokenExpiry"`
	Status       PlatformStatus `json:"status"`
	Metadata     JSON           `json:"metadata" gorm:"type:jsonb"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

// ScheduledPost represents a post scheduled for publishing
type ScheduledPost struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	UserID      string         `json:"userId" gorm:"index"`
	VideoID     string         `json:"videoId"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Platforms   []PlatformPost `json:"platforms" gorm:"-"`
	ScheduledAt time.Time      `json:"scheduledAt"`
	Timezone    string         `json:"timezone"`
	Status      PostStatus     `json:"status"`
	Recurring   *RecurringRule `json:"recurring,omitempty" gorm:"-"`
	Metadata    JSON           `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	PublishedAt *time.Time     `json:"publishedAt"`
	ErrorMsg    string         `json:"errorMsg,omitempty"`
}

// PlatformPost represents a post configuration for a specific platform
type PlatformPost struct {
	ID              string         `json:"id" gorm:"primaryKey"`
	ScheduledPostID string         `json:"scheduledPostId" gorm:"index"`
	Platform        SocialPlatform `json:"platform"`
	AccountID       string         `json:"accountId"`
	Status          PostStatus     `json:"status"`
	PlatformPostID  string         `json:"platformPostId,omitempty"`
	PostURL         string         `json:"postUrl,omitempty"`
	CustomTitle     string         `json:"customTitle,omitempty"`
	CustomDesc      string         `json:"customDesc,omitempty"`
	Tags            []string       `json:"tags,omitempty" gorm:"-"`
	Privacy         string         `json:"privacy,omitempty"`
	Metadata        JSON           `json:"metadata" gorm:"type:jsonb"`
	ErrorMsg        string         `json:"errorMsg,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	PublishedAt     *time.Time     `json:"publishedAt"`
}

// RecurringRule defines how a post should repeat
type RecurringRule struct {
	Frequency  string   `json:"frequency"` // daily, weekly, monthly
	Interval   int      `json:"interval"`  // every N days/weeks/months
	DaysOfWeek []int    `json:"daysOfWeek,omitempty"`
	EndDate    *string  `json:"endDate,omitempty"`
	EndAfter   *int     `json:"endAfter,omitempty"`
}

// AnalyticsData represents platform analytics for a post
type AnalyticsData struct {
	ID           string         `json:"id" gorm:"primaryKey"`
	PostID       string         `json:"postId" gorm:"index"`
	Platform     SocialPlatform `json:"platform"`
	Views        int64          `json:"views"`
	Likes        int64          `json:"likes"`
	Comments     int64          `json:"comments"`
	Shares       int64          `json:"shares"`
	WatchTime    int64          `json:"watchTime"` // in seconds
	Subscribers  int64          `json:"subscribers"`
	Engagement   float64        `json:"engagement"`
	Data         JSON           `json:"data" gorm:"type:jsonb"`
	RecordedAt   time.Time      `json:"recordedAt"`
}

// PlatformTrend represents trending topics/sounds for a platform
type PlatformTrend struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	Platform    SocialPlatform `json:"platform"`
	TrendType   string         `json:"trendType"` // hashtag, sound, challenge
	Title       string         `json:"title"`
	Description string         `json:"description"`
	URL         string         `json:"url,omitempty"`
	Volume      int64          `json:"volume"`
	Region      string         `json:"region"`
	FetchedAt   time.Time      `json:"fetchedAt"`
}

// UploadRequest represents a video upload request
type UploadRequest struct {
	VideoPath   string            `json:"videoPath"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Privacy     string            `json:"privacy"` // public, unlisted, private
	Metadata    map[string]string `json:"metadata"`
}

// UploadResponse represents the result of an upload
type UploadResponse struct {
	PlatformPostID string `json:"platformPostId"`
	PostURL        string `json:"postUrl"`
	Status         string `json:"status"`
}

// JSON is a custom type for JSONB fields
type JSON map[string]interface{}
