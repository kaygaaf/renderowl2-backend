package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"renderowl-api/internal/domain"
	"renderowl-api/internal/repository"
)

// AnalyticsService handles analytics business logic
type AnalyticsService struct {
	analyticsRepo *repository.AnalyticsRepository
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(analyticsRepo *repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
	}
}

// TrackViewRequest represents a view tracking request
type TrackViewRequest struct {
	VideoID  string `json:"video_id" binding:"required"`
	Platform string `json:"platform" binding:"required"`
	UserID   string `json:"user_id"`
}

// TrackView records a video view
func (s *AnalyticsService) TrackView(ctx context.Context, req *TrackViewRequest) error {
	return s.analyticsRepo.TrackView(ctx, req.VideoID, req.UserID, req.Platform)
}

// ViewsOverTimeResponse represents views over time data
type ViewsOverTimeResponse struct {
	Data []DailyViews `json:"data"`
}

// DailyViews represents views for a single day
type DailyViews struct {
	Date     string         `json:"date"`
	Total    int64          `json:"total"`
	ByPlatform map[string]int64 `json:"by_platform"`
}

// GetViewsOverTime gets views over time for a user
func (s *AnalyticsService) GetViewsOverTime(ctx context.Context, userID string, days int) (*ViewsOverTimeResponse, error) {
	startDate := time.Now().UTC().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	endDate := time.Now().UTC().Truncate(24 * time.Hour)
	
	aggregates, err := s.analyticsRepo.GetViewsByDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	
	// Group by date
	dateMap := make(map[string]*DailyViews)
	for _, agg := range aggregates {
		dateStr := agg.Date.Format("2006-01-02")
		if _, exists := dateMap[dateStr]; !exists {
			dateMap[dateStr] = &DailyViews{
				Date:       dateStr,
				ByPlatform: make(map[string]int64),
			}
		}
		dateMap[dateStr].Total += agg.TotalViews
		dateMap[dateStr].ByPlatform[agg.Platform] = agg.TotalViews
	}
	
	// Convert to slice and fill missing dates
	result := make([]DailyViews, 0, days)
	for i := 0; i < days; i++ {
		date := time.Now().UTC().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		dateStr := date.Format("2006-01-02")
		
		if data, exists := dateMap[dateStr]; exists {
			result = append(result, *data)
		} else {
			result = append(result, DailyViews{
				Date:       dateStr,
				Total:      0,
				ByPlatform: make(map[string]int64),
			})
		}
	}
	
	return &ViewsOverTimeResponse{Data: result}, nil
}

// PlatformBreakdownResponse represents platform breakdown data
type PlatformBreakdownResponse struct {
	Platforms []PlatformData `json:"platforms"`
}

// PlatformData represents data for a single platform
type PlatformData struct {
	Platform    string  `json:"platform"`
	Views       int64   `json:"views"`
	Percentage  float64 `json:"percentage"`
	Videos      int64   `json:"videos"`
	Followers   int64   `json:"followers"`
}

// GetPlatformBreakdown gets platform breakdown for a user
func (s *AnalyticsService) GetPlatformBreakdown(ctx context.Context, userID string, days int) (*PlatformBreakdownResponse, error) {
	platformViews, err := s.analyticsRepo.GetViewsByPlatform(ctx, userID, days)
	if err != nil {
		return nil, err
	}
	
	// Calculate total for percentages
	var totalViews int64
	for _, pv := range platformViews {
		totalViews += pv.Views
	}
	
	// Build response
	platforms := make([]PlatformData, 0, len(platformViews))
	for _, pv := range platformViews {
		percentage := 0.0
		if totalViews > 0 {
			percentage = float64(pv.Views) / float64(totalViews) * 100
		}
		
		platforms = append(platforms, PlatformData{
			Platform:   pv.Platform,
			Views:      pv.Views,
			Percentage: percentage,
		})
	}
	
	return &PlatformBreakdownResponse{Platforms: platforms}, nil
}

// EngagementMetricsResponse represents engagement metrics
type EngagementMetricsResponse struct {
	TotalLikes        int64              `json:"total_likes"`
	TotalComments     int64              `json:"total_comments"`
	TotalShares       int64              `json:"total_shares"`
	TotalSaves        int64              `json:"total_saves"`
	EngagementRate    float64            `json:"engagement_rate"`
	ByPlatform        map[string]PlatformEngagement `json:"by_platform"`
}

// PlatformEngagement represents engagement for a platform
type PlatformEngagement struct {
	Likes    int64   `json:"likes"`
	Comments int64   `json:"comments"`
	Shares   int64   `json:"shares"`
	Rate     float64 `json:"rate"`
}

// GetEngagementMetrics gets engagement metrics for a user
func (s *AnalyticsService) GetEngagementMetrics(ctx context.Context, userID string, days int) (*EngagementMetricsResponse, error) {
	engagementRate, err := s.analyticsRepo.GetEngagementRate(ctx, userID, days)
	if err != nil {
		return nil, err
	}
	
	return &EngagementMetricsResponse{
		EngagementRate: engagementRate,
		ByPlatform:     make(map[string]PlatformEngagement),
	}, nil
}

// TrackEngagementRequest represents an engagement tracking request
type TrackEngagementRequest struct {
	VideoID   string `json:"video_id" binding:"required"`
	Platform  string `json:"platform" binding:"required"`
	Likes     int64  `json:"likes"`
	Comments  int64  `json:"comments"`
	Shares    int64  `json:"shares"`
	Saves     int64  `json:"saves"`
}

// TrackEngagement records engagement metrics
func (s *AnalyticsService) TrackEngagement(ctx context.Context, req *TrackEngagementRequest) error {
	return s.analyticsRepo.TrackEngagement(ctx, req.VideoID, req.Platform, req.Likes, req.Comments, req.Shares)
}

// UserGrowthResponse represents user growth data
type UserGrowthResponse struct {
	Data []DailyGrowth `json:"data"`
}

// DailyGrowth represents growth for a single day
type DailyGrowth struct {
	Date           string `json:"date"`
	NewSignups     int64  `json:"new_signups"`
	ActiveUsers    int64  `json:"active_users"`
	ReturningUsers int64  `json:"returning_users"`
}

// GetUserGrowth gets user growth data
func (s *AnalyticsService) GetUserGrowth(ctx context.Context, days int) (*UserGrowthResponse, error) {
	growthData, err := s.analyticsRepo.GetUserGrowth(ctx, days)
	if err != nil {
		return nil, err
	}
	
	result := make([]DailyGrowth, 0, len(growthData))
	for _, g := range growthData {
		result = append(result, DailyGrowth{
			Date:           g.Date.Format("2006-01-02"),
			NewSignups:     g.NewSignups,
			ActiveUsers:    g.ActiveUsers,
			ReturningUsers: g.ReturningUsers,
		})
	}
	
	return &UserGrowthResponse{Data: result}, nil
}

// RevenueReportResponse represents revenue report data
type RevenueReportResponse struct {
	TotalRevenue      float64           `json:"total_revenue"`
	TotalTransactions int64             `json:"total_transactions"`
	ByType            map[string]float64 `json:"by_type"`
	Daily             []DailyRevenue    `json:"daily"`
}

// DailyRevenue represents revenue for a single day
type DailyRevenue struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

// GetRevenueReport gets revenue report
func (s *AnalyticsService) GetRevenueReport(ctx context.Context, days int) (*RevenueReportResponse, error) {
	summary, err := s.analyticsRepo.GetRevenueSummary(ctx, days)
	if err != nil {
		return nil, err
	}
	
	byType := make(map[string]float64)
	for _, t := range summary.ByType {
		byType[t.Type] = t.Amount
	}
	
	return &RevenueReportResponse{
		TotalRevenue:      summary.TotalRevenue,
		TotalTransactions: summary.TotalTransactions,
		ByType:            byType,
		Daily:             []DailyRevenue{},
	}, nil
}

// RecordRevenueRequest represents a revenue recording request
type RecordRevenueRequest struct {
	UserID      string  `json:"user_id"`
	Type        string  `json:"type" binding:"required"` // subscription, credits, one_time
	Amount      float64 `json:"amount" binding:"required"`
	Currency    string  `json:"currency"`
	Platform    string  `json:"platform"` // stripe, paypal
	Description string  `json:"description"`
}

// RecordRevenue records a revenue transaction
func (s *AnalyticsService) RecordRevenue(ctx context.Context, req *RecordRevenueRequest) error {
	revenue := &domain.Revenue{
		UserID:      req.UserID,
		Type:        req.Type,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Platform:    req.Platform,
		Date:        time.Now().UTC(),
		Description: req.Description,
	}
	
	if revenue.Currency == "" {
		revenue.Currency = "USD"
	}
	if revenue.Platform == "" {
		revenue.Platform = "stripe"
	}
	
	return s.analyticsRepo.RecordRevenue(ctx, revenue)
}

// VideoPerformanceResponse represents video performance data
type VideoPerformanceResponse struct {
	Videos     []VideoMetrics `json:"videos"`
	TotalCount int64          `json:"total_count"`
}

// VideoMetrics represents metrics for a single video
type VideoMetrics struct {
	VideoID        string   `json:"video_id"`
	Title          string   `json:"title"`
	Thumbnail      string   `json:"thumbnail,omitempty"`
	Views          int64    `json:"views"`
	Likes          int64    `json:"likes"`
	Comments       int64    `json:"comments"`
	Shares         int64    `json:"shares"`
	EngagementRate float64  `json:"engagement_rate"`
	Platforms      []string `json:"platforms"`
	PublishedAt    *string  `json:"published_at,omitempty"`
	Duration       float64  `json:"duration,omitempty"`
}

// GetVideoPerformance gets video performance for a user
func (s *AnalyticsService) GetVideoPerformance(ctx context.Context, userID string, limit, offset int) (*VideoPerformanceResponse, error) {
	performanceData, err := s.analyticsRepo.GetVideoPerformance(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	
	videos := make([]VideoMetrics, 0, len(performanceData))
	for _, p := range performanceData {
		var publishedAt *string
		if p.PublishedAt != nil {
			str := p.PublishedAt.Format("2006-01-02")
			publishedAt = &str
		}
		
		videos = append(videos, VideoMetrics{
			VideoID:        p.VideoID,
			Title:          p.Title,
			Views:          p.TotalViews,
			Likes:          p.TotalLikes,
			Comments:       p.TotalComments,
			Shares:         p.TotalShares,
			EngagementRate: p.EngagementRate,
			Platforms:      p.Platforms,
			PublishedAt:    publishedAt,
		})
	}
	
	return &VideoPerformanceResponse{
		Videos:     videos,
		TotalCount: int64(len(videos)),
	}, nil
}

// UpdateVideoPerformanceRequest represents a video performance update request
type UpdateVideoPerformanceRequest struct {
	VideoID        string   `json:"video_id" binding:"required"`
	UserID         string   `json:"user_id" binding:"required"`
	Title          string   `json:"title"`
	TotalViews     int64    `json:"total_views"`
	TotalLikes     int64    `json:"total_likes"`
	TotalComments  int64    `json:"total_comments"`
	TotalShares    int64    `json:"total_shares"`
	Platforms      []string `json:"platforms"`
}

// UpdateVideoPerformance updates video performance metrics
func (s *AnalyticsService) UpdateVideoPerformance(ctx context.Context, req *UpdateVideoPerformanceRequest) error {
	// Calculate engagement rate
	var engagementRate float64
	if req.TotalViews > 0 {
		engagementRate = float64(req.TotalLikes+req.TotalComments+req.TotalShares) / float64(req.TotalViews) * 100
	}
	
	performance := &domain.VideoPerformance{
		VideoID:        req.VideoID,
		UserID:         req.UserID,
		Title:          req.Title,
		TotalViews:     req.TotalViews,
		TotalLikes:     req.TotalLikes,
		TotalComments:  req.TotalComments,
		TotalShares:    req.TotalShares,
		EngagementRate: engagementRate,
		Platforms:      req.Platforms,
	}
	
	return s.analyticsRepo.UpdateVideoPerformance(ctx, performance)
}

// DashboardSummaryResponse represents the dashboard summary
type DashboardSummaryResponse struct {
	TotalViews       int64                `json:"total_views"`
	TotalEngagements int64                `json:"total_engagements"`
	ViewsLast30Days  int64                `json:"views_last_30_days"`
	EngagementRate   float64              `json:"engagement_rate"`
	TotalVideos      int64                `json:"total_videos"`
	RevenueLast30Days float64             `json:"revenue_last_30_days"`
	TopVideo         *VideoMetrics        `json:"top_video,omitempty"`
	PlatformBreakdown []PlatformData      `json:"platform_breakdown"`
}

// GetDashboardSummary gets the main dashboard summary
func (s *AnalyticsService) GetDashboardSummary(ctx context.Context, userID string) (*DashboardSummaryResponse, error) {
	summary, err := s.analyticsRepo.GetDashboardSummary(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Get platform breakdown
	platformBreakdown, err := s.GetPlatformBreakdown(ctx, userID, 30)
	if err != nil {
		platformBreakdown = &PlatformBreakdownResponse{Platforms: []PlatformData{}}
	}
	
	// Get revenue for last 30 days
	revenueSummary, err := s.analyticsRepo.GetRevenueSummary(ctx, 30)
	if err != nil {
		revenueSummary = &repository.RevenueSummary{}
	}
	
	response := &DashboardSummaryResponse{
		TotalViews:        summary.TotalViews,
		TotalEngagements:  summary.TotalEngagements,
		ViewsLast30Days:   summary.ViewsLast30Days,
		EngagementRate:    summary.EngagementRate,
		RevenueLast30Days: revenueSummary.TotalRevenue,
		PlatformBreakdown: platformBreakdown.Platforms,
	}
	
	// Convert top video
	if summary.TopVideo != nil {
		var publishedAt *string
		if summary.TopVideo.PublishedAt != nil {
			str := summary.TopVideo.PublishedAt.Format("2006-01-02")
			publishedAt = &str
		}
		
		response.TopVideo = &VideoMetrics{
			VideoID:        summary.TopVideo.VideoID,
			Title:          summary.TopVideo.Title,
			Views:          summary.TopVideo.TotalViews,
			Likes:          summary.TopVideo.TotalLikes,
			Comments:       summary.TopVideo.TotalComments,
			Shares:         summary.TopVideo.TotalShares,
			EngagementRate: summary.TopVideo.EngagementRate,
			PublishedAt:    publishedAt,
		}
	}
	
	return response, nil
}

// AnalyticsOverviewResponse represents the analytics overview
type AnalyticsOverviewResponse struct {
	Period            string             `json:"period"`
	TotalViews        int64              `json:"total_views"`
	ViewsByPlatform   map[string]int64   `json:"views_by_platform"`
	TotalVideos       int64              `json:"total_videos"`
	AvgEngagementRate float64            `json:"avg_engagement_rate"`
	DailyViews        []DailyViews       `json:"daily_views"`
	TopVideos         []VideoMetrics     `json:"top_videos"`
}

// GetAnalyticsOverview gets comprehensive analytics overview
func (s *AnalyticsService) GetAnalyticsOverview(ctx context.Context, userID string, days int) (*AnalyticsOverviewResponse, error) {
	// Get basic overview
	overview, err := s.analyticsRepo.GetAnalyticsOverview(ctx, userID, days)
	if err != nil {
		return nil, err
	}
	
	// Get views over time
	viewsOverTime, err := s.GetViewsOverTime(ctx, userID, days)
	if err != nil {
		viewsOverTime = &ViewsOverTimeResponse{Data: []DailyViews{}}
	}
	
	// Get top videos
	videoPerformance, err := s.GetVideoPerformance(ctx, userID, 10, 0)
	if err != nil {
		videoPerformance = &VideoPerformanceResponse{Videos: []VideoMetrics{}}
	}
	
	period := fmt.Sprintf("Last %d days", days)
	if days == 7 {
		period = "Last 7 days"
	} else if days == 30 {
		period = "Last 30 days"
	}
	
	return &AnalyticsOverviewResponse{
		Period:            period,
		TotalViews:        overview.TotalViews,
		ViewsByPlatform:   overview.ViewsByPlatform,
		TotalVideos:       overview.TotalVideos,
		AvgEngagementRate: overview.AvgEngagementRate,
		DailyViews:        viewsOverTime.Data,
		TopVideos:         videoPerformance.Videos,
	}, nil
}

// WebhookEventRequest represents a webhook event request
type WebhookEventRequest struct {
	Platform  string                 `json:"platform" binding:"required"` // youtube, tiktok, instagram
	EventType string                 `json:"event_type" binding:"required"`
	VideoID   string                 `json:"video_id"`
	Payload   map[string]interface{} `json:"payload"`
}

// StoreWebhookEvent stores a webhook event
func (s *AnalyticsService) StoreWebhookEvent(ctx context.Context, req *WebhookEventRequest) (string, error) {
	event := &domain.WebhookEvent{
		Platform:  req.Platform,
		EventType: req.EventType,
		VideoID:   req.VideoID,
		Payload:   req.Payload,
	}
	
	err := s.analyticsRepo.StoreWebhookEvent(ctx, event)
	if err != nil {
		return "", err
	}
	
	return event.ID, nil
}

// ProcessWebhookEvents processes unprocessed webhook events
func (s *AnalyticsService) ProcessWebhookEvents(ctx context.Context, limit int) error {
	events, err := s.analyticsRepo.GetUnprocessedWebhookEvents(ctx, limit)
	if err != nil {
		return err
	}
	
	for _, event := range events {
		// Process based on event type and platform
		switch event.EventType {
		case "view":
			if videoID, ok := event.Payload["video_id"].(string); ok {
				if count, ok := event.Payload["count"].(float64); ok {
					s.analyticsRepo.TrackView(ctx, videoID, "", event.Platform)
					_ = count // Use the count value
				}
			}
		case "engagement":
			if videoID, ok := event.Payload["video_id"].(string); ok {
				likes := int64(0)
				comments := int64(0)
				shares := int64(0)
				
				if v, ok := event.Payload["likes"].(float64); ok {
					likes = int64(v)
				}
				if v, ok := event.Payload["comments"].(float64); ok {
					comments = int64(v)
				}
				if v, ok := event.Payload["shares"].(float64); ok {
					shares = int64(v)
				}
				
				s.analyticsRepo.TrackEngagement(ctx, videoID, event.Platform, likes, comments, shares)
			}
		}
		
		// Mark as processed
		s.analyticsRepo.MarkWebhookEventProcessed(ctx, event.ID)
	}
	
	return nil
}

// ExportAnalyticsRequest represents an analytics export request
type ExportAnalyticsRequest struct {
	UserID    string    `json:"user_id" binding:"required"`
	StartDate time.Time `json:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date" binding:"required"`
	Format    string    `json:"format"` // csv, json
}

// ExportAnalytics exports analytics data
func (s *AnalyticsService) ExportAnalytics(ctx context.Context, req *ExportAnalyticsRequest) ([]byte, error) {
	if req.Format == "" {
		req.Format = "json"
	}
	
	// Get views data
	views, err := s.analyticsRepo.GetViewsByDateRange(ctx, req.UserID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	
	// For now, return JSON format
	if req.Format == "json" {
		return json.Marshal(map[string]interface{}{
			"period": map[string]string{
				"start": req.StartDate.Format("2006-01-02"),
				"end":   req.EndDate.Format("2006-01-02"),
			},
			"views": views,
		})
	}
	
	return nil, fmt.Errorf("unsupported format: %s", req.Format)
}
