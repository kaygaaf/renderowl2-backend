package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"renderowl-api/internal/domain"
)

// AnalyticsRepository handles analytics database operations
type AnalyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// TrackView records a video view
func (r *AnalyticsRepository) TrackView(ctx context.Context, videoID, userID, platform string) error {
	view := domain.AnalyticsView{
		VideoID:  videoID,
		UserID:   userID,
		Platform: platform,
		Date:     time.Now().UTC().Truncate(24 * time.Hour),
	}
	return r.db.WithContext(ctx).Create(&view).Error
}

// GetViewsByDateRange gets views for a date range
func (r *AnalyticsRepository) GetViewsByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]ViewAggregate, error) {
	var results []ViewAggregate
	
	err := r.db.WithContext(ctx).Model(&domain.AnalyticsView{}).
		Select("date, platform, SUM(count) as total_views").
		Where("date >= ? AND date <= ?", startDate, endDate).
		Group("date, platform").
		Order("date DESC").
		Find(&results).Error
	
	return results, err
}

// ViewAggregate represents aggregated view data
type ViewAggregate struct {
	Date        time.Time `json:"date"`
	Platform    string    `json:"platform"`
	TotalViews  int64     `json:"total_views"`
}

// GetTotalViews gets total views for a user
func (r *AnalyticsRepository) GetTotalViews(ctx context.Context, userID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&domain.AnalyticsView{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(count), 0)").
		Scan(&total).Error
	return total, err
}

// GetViewsByPlatform gets view counts grouped by platform
func (r *AnalyticsRepository) GetViewsByPlatform(ctx context.Context, userID string, days int) ([]PlatformViews, error) {
	var results []PlatformViews
	startDate := time.Now().UTC().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	
	err := r.db.WithContext(ctx).Model(&domain.AnalyticsView{}).
		Select("platform, SUM(count) as views").
		Where("user_id = ? AND date >= ?", userID, startDate).
		Group("platform").
		Find(&results).Error
	
	return results, err
}

// PlatformViews represents platform view counts
type PlatformViews struct {
	Platform string `json:"platform"`
	Views    int64  `json:"views"`
}

// TrackEngagement records engagement metrics
func (r *AnalyticsRepository) TrackEngagement(ctx context.Context, videoID, platform string, likes, comments, shares int64) error {
	engagement := domain.AnalyticsEngagement{
		VideoID:  videoID,
		Platform: platform,
		Likes:    likes,
		Comments: comments,
		Shares:   shares,
		Date:     time.Now().UTC().Truncate(24 * time.Hour),
	}
	
	// Upsert: update if exists for this video/platform/date
	return r.db.WithContext(ctx).Where(
		"video_id = ? AND platform = ? AND date = ?",
		videoID, platform, engagement.Date,
	).Assign(domain.AnalyticsEngagement{
		Likes:    likes,
		Comments: comments,
		Shares:   shares,
	}).FirstOrCreate(&engagement).Error
}

// GetEngagementByVideo gets engagement metrics for a video
func (r *AnalyticsRepository) GetEngagementByVideo(ctx context.Context, videoID string) (*EngagementSummary, error) {
	var result EngagementSummary
	
	err := r.db.WithContext(ctx).Model(&domain.AnalyticsEngagement{}).
		Select("COALESCE(SUM(likes), 0) as total_likes, COALESCE(SUM(comments), 0) as total_comments, COALESCE(SUM(shares), 0) as total_shares").
		Where("video_id = ?", videoID).
		Scan(&result).Error
	
	return &result, err
}

// EngagementSummary represents aggregated engagement data
type EngagementSummary struct {
	TotalLikes    int64 `json:"total_likes"`
	TotalComments int64 `json:"total_comments"`
	TotalShares   int64 `json:"total_shares"`
}

// GetEngagementRate calculates engagement rate for a user
func (r *AnalyticsRepository) GetEngagementRate(ctx context.Context, userID string, days int) (float64, error) {
	startDate := time.Now().UTC().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	
	var views int64
	var likes int64
	
	// Get total views
	r.db.WithContext(ctx).Model(&domain.AnalyticsView{}).
		Where("user_id = ? AND date >= ?", userID, startDate).
		Select("COALESCE(SUM(count), 0)").
		Scan(&views)
	
	// Get total engagement
	r.db.WithContext(ctx).Model(&domain.AnalyticsEngagement{}).
		Joins("JOIN analytics_views ON analytics_engagement.video_id = analytics_views.video_id").
		Where("analytics_views.user_id = ? AND analytics_engagement.date >= ?", userID, startDate).
		Select("COALESCE(SUM(likes + comments + shares), 0)").
		Scan(&likes)
	
	if views == 0 {
		return 0, nil
	}
	
	return float64(likes) / float64(views) * 100, nil
}

// RecordUserSignup records a new user signup
func (r *AnalyticsRepository) RecordUserSignup(ctx context.Context, date time.Time) error {
	growth := domain.UserGrowth{Date: date.Truncate(24 * time.Hour)}
	
	return r.db.WithContext(ctx).Model(&domain.UserGrowth{}).
		Where("date = ?", growth.Date).
		UpdateColumn("new_signups", gorm.Expr("new_signups + 1")).
		Error
}

// GetUserGrowth gets user growth data for a date range
func (r *AnalyticsRepository) GetUserGrowth(ctx context.Context, days int) ([]UserGrowthData, error) {
	var results []UserGrowthData
	startDate := time.Now().UTC().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	
	err := r.db.WithContext(ctx).Model(&domain.UserGrowth{}).
		Select("date, new_signups, active_users, returning_users").
		Where("date >= ?", startDate).
		Order("date ASC").
		Find(&results).Error
	
	return results, err
}

// UserGrowthData represents user growth metrics
type UserGrowthData struct {
	Date           time.Time `json:"date"`
	NewSignups     int64     `json:"new_signups"`
	ActiveUsers    int64     `json:"active_users"`
	ReturningUsers int64     `json:"returning_users"`
}

// RecordRevenue records a revenue transaction
func (r *AnalyticsRepository) RecordRevenue(ctx context.Context, revenue *domain.Revenue) error {
	return r.db.WithContext(ctx).Create(revenue).Error
}

// GetRevenueSummary gets revenue summary for a date range
func (r *AnalyticsRepository) GetRevenueSummary(ctx context.Context, days int) (*RevenueSummary, error) {
	startDate := time.Now().UTC().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	
	var result RevenueSummary
	err := r.db.WithContext(ctx).Model(&domain.Revenue{}).
		Select("COALESCE(SUM(amount), 0) as total_revenue, COUNT(*) as total_transactions").
		Where("date >= ? AND status = 'completed'", startDate).
		Scan(&result).Error
	
	// Get revenue by type
	var byType []RevenueByType
	r.db.WithContext(ctx).Model(&domain.Revenue{}).
		Select("type, COALESCE(SUM(amount), 0) as amount").
		Where("date >= ? AND status = 'completed'", startDate).
		Group("type").
		Find(&byType)
	
	result.ByType = byType
	return &result, err
}

// RevenueSummary represents revenue aggregation
type RevenueSummary struct {
	TotalRevenue      float64         `json:"total_revenue"`
	TotalTransactions int64           `json:"total_transactions"`
	ByType            []RevenueByType `json:"by_type"`
}

// RevenueByType represents revenue grouped by type
type RevenueByType struct {
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}

// GetVideoPerformance gets performance data for all user videos
func (r *AnalyticsRepository) GetVideoPerformance(ctx context.Context, userID string, limit, offset int) ([]VideoPerformanceData, error) {
	var results []VideoPerformanceData
	
	err := r.db.WithContext(ctx).Model(&domain.VideoPerformance{}).
		Where("user_id = ?", userID).
		Order("total_views DESC").
		Limit(limit).Offset(offset).
		Find(&results).Error
	
	return results, err
}

// VideoPerformanceData represents video performance metrics
type VideoPerformanceData struct {
	VideoID        string    `json:"video_id"`
	Title          string    `json:"title"`
	TotalViews     int64     `json:"total_views"`
	TotalLikes     int64     `json:"total_likes"`
	TotalComments  int64     `json:"total_comments"`
	TotalShares    int64     `json:"total_shares"`
	EngagementRate float64   `json:"engagement_rate"`
	Platforms      []string  `json:"platforms"`
	PublishedAt    *time.Time `json:"published_at"`
}

// UpdateVideoPerformance updates or creates video performance record
func (r *AnalyticsRepository) UpdateVideoPerformance(ctx context.Context, performance *domain.VideoPerformance) error {
	performance.LastUpdated = time.Now().UTC()
	
	return r.db.WithContext(ctx).Where(
		"video_id = ?", performance.VideoID,
	).Assign(domain.VideoPerformance{
		TotalViews:     performance.TotalViews,
		TotalLikes:     performance.TotalLikes,
		TotalComments:  performance.TotalComments,
		TotalShares:    performance.TotalShares,
		EngagementRate: performance.EngagementRate,
		LastUpdated:    performance.LastUpdated,
	}).FirstOrCreate(performance).Error
}

// GetPlatformStats gets aggregated stats for all platforms
func (r *AnalyticsRepository) GetPlatformStats(ctx context.Context) ([]PlatformStatData, error) {
	var results []PlatformStatData
	
	err := r.db.WithContext(ctx).Model(&domain.PlatformStats{}).
		Select("platform, total_views, total_videos, total_likes, followers").
		Find(&results).Error
	
	return results, err
}

// PlatformStatData represents platform statistics
type PlatformStatData struct {
	Platform    string `json:"platform"`
	TotalViews  int64  `json:"total_views"`
	TotalVideos int64  `json:"total_videos"`
	TotalLikes  int64  `json:"total_likes"`
	Followers   int64  `json:"followers"`
}

// StoreWebhookEvent stores an incoming webhook event
func (r *AnalyticsRepository) StoreWebhookEvent(ctx context.Context, event *domain.WebhookEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// GetUnprocessedWebhookEvents gets unprocessed webhook events
func (r *AnalyticsRepository) GetUnprocessedWebhookEvents(ctx context.Context, limit int) ([]domain.WebhookEvent, error) {
	var events []domain.WebhookEvent
	
	err := r.db.WithContext(ctx).Where("processed = ?", false).
		Limit(limit).
		Find(&events).Error
	
	return events, err
}

// MarkWebhookEventProcessed marks a webhook event as processed
func (r *AnalyticsRepository) MarkWebhookEventProcessed(ctx context.Context, eventID string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&domain.WebhookEvent{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"processed":    true,
			"processed_at": now,
		}).Error
}

// GetDashboardSummary gets comprehensive dashboard summary
func (r *AnalyticsRepository) GetDashboardSummary(ctx context.Context, userID string) (*DashboardSummary, error) {
	summary := &DashboardSummary{}
	
	// Get total views
	r.db.WithContext(ctx).Model(&domain.AnalyticsView{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(count), 0)").
		Scan(&summary.TotalViews)
	
	// Get total engagement
	r.db.WithContext(ctx).Model(&domain.AnalyticsEngagement{}).
		Joins("JOIN analytics_views ON analytics_engagement.video_id = analytics_views.video_id").
		Where("analytics_views.user_id = ?", userID).
		Select("COALESCE(SUM(likes + comments + shares), 0)").
		Scan(&summary.TotalEngagements)
	
	// Get views in last 30 days
	startDate := time.Now().UTC().AddDate(0, 0, -30).Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&domain.AnalyticsView{}).
		Where("user_id = ? AND date >= ?", userID, startDate).
		Select("COALESCE(SUM(count), 0)").
		Scan(&summary.ViewsLast30Days)
	
	// Calculate engagement rate
	if summary.TotalViews > 0 {
		summary.EngagementRate = float64(summary.TotalEngagements) / float64(summary.TotalViews) * 100
	}
	
	// Get top performing video
	var topVideo VideoPerformanceData
	r.db.WithContext(ctx).Model(&domain.VideoPerformance{}).
		Where("user_id = ?", userID).
		Order("total_views DESC").
		Limit(1).
		Find(&topVideo)
	summary.TopVideo = &topVideo
	
	return summary, nil
}

// DashboardSummary represents the main dashboard overview
type DashboardSummary struct {
	TotalViews        int64                 `json:"total_views"`
	TotalEngagements  int64                 `json:"total_engagements"`
	ViewsLast30Days   int64                 `json:"views_last_30_days"`
	EngagementRate    float64               `json:"engagement_rate"`
	TopVideo          *VideoPerformanceData `json:"top_video,omitempty"`
}

// GetAnalyticsOverview gets the main analytics overview data
func (r *AnalyticsRepository) GetAnalyticsOverview(ctx context.Context, userID string, days int) (*AnalyticsOverview, error) {
	overview := &AnalyticsOverview{}
	startDate := time.Now().UTC().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	
	// Total views in period
	r.db.WithContext(ctx).Model(&domain.AnalyticsView{}).
		Where("user_id = ? AND date >= ?", userID, startDate).
		Select("COALESCE(SUM(count), 0)").
		Scan(&overview.TotalViews)
	
	// Views by platform
	var platformData []struct {
		Platform string
		Views    int64
	}
	r.db.WithContext(ctx).Model(&domain.AnalyticsView{}).
		Select("platform, SUM(count) as views").
		Where("user_id = ? AND date >= ?", userID, startDate).
		Group("platform").
		Find(&platformData)
	
	overview.ViewsByPlatform = make(map[string]int64)
	for _, p := range platformData {
		overview.ViewsByPlatform[p.Platform] = p.Views
	}
	
	// Total videos published
	r.db.WithContext(ctx).Model(&domain.VideoPerformance{}).
		Where("user_id = ? AND published_at >= ?", userID, startDate).
		Select("COUNT(*)").
		Scan(&overview.TotalVideos)
	
	// Average engagement rate
	r.db.WithContext(ctx).Model(&domain.VideoPerformance{}).
		Where("user_id = ?", userID).
		Select("COALESCE(AVG(engagement_rate), 0)").
		Scan(&overview.AvgEngagementRate)
	
	return overview, nil
}

// AnalyticsOverview represents the analytics overview response
type AnalyticsOverview struct {
	TotalViews        int64            `json:"total_views"`
	ViewsByPlatform   map[string]int64 `json:"views_by_platform"`
	TotalVideos       int64            `json:"total_videos"`
	AvgEngagementRate float64          `json:"avg_engagement_rate"`
}
