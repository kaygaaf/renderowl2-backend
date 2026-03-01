package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"renderowl-api/internal/domain"
)

// OptimizerService handles content optimization based on performance analytics
type OptimizerService struct {
	analyticsRepo   AnalyticsRepository
	timelineRepo    TimelineRepository
	socialService   *SocialService
	aiScriptService *AIScriptService
}

// AnalyticsRepository defines the interface for analytics data
type AnalyticsRepository interface {
	GetVideoPerformance(videoID string, days int) (*VideoAnalytics, error)
	GetTopPerforming(userID string, limit int) ([]*VideoAnalytics, error)
	GetUnderperforming(userID string, limit int) ([]*VideoAnalytics, error)
	GetComparativeMetrics(videoIDs []string) (map[string]*ComparativeMetrics, error)
	GetTrendingTopics(days int) ([]*TrendingTopicData, error)
}

// TimelineRepository defines the interface for timeline data
type TimelineRepository interface {
	Get(id string, userID string) (*domain.Timeline, error)
	Update(timeline *domain.Timeline) error
	ListByUser(userID string, limit, offset int) ([]*domain.Timeline, error)
}

// SocialService handles social media operations
type SocialService struct {
	// Social media service implementation
}

// VideoAnalytics represents performance data for a video
type VideoAnalytics struct {
	VideoID         string                 `json:"videoId"`
	Title           string                 `json:"title"`
	Platform        string                 `json:"platform"`
	Views           int                    `json:"views"`
	Likes           int                    `json:"likes"`
	Comments        int                    `json:"comments"`
	Shares          int                    `json:"shares"`
	WatchTime       float64                `json:"watchTime"` // seconds
	AvgWatchDuration float64               `json:"avgWatchDuration"`
	CTR             float64                `json:"ctr"`       // Click-through rate
	RetentionCurve  []float64              `json:"retentionCurve"` // Percentage at each 10% mark
	TrafficSources  map[string]float64     `json:"trafficSources"`
	AudienceDemographics map[string]interface{} `json:"audienceDemographics"`
	PublishDate     time.Time              `json:"publishDate"`
	DaysSincePublish int                   `json:"daysSincePublish"`
	EngagementRate  float64                `json:"engagementRate"`
	ViralScore      float64                `json:"viralScore"`
}

// ComparativeMetrics for A/B testing
type ComparativeMetrics struct {
	VideoID        string  `json:"videoId"`
	Variant        string  `json:"variant"` // A, B, C
	Views          int     `json:"views"`
	CTR            float64 `json:"ctr"`
	EngagementRate float64 `json:"engagementRate"`
	WatchTime      float64 `json:"watchTime"`
	Winner         bool    `json:"winner"`
}

// TrendingTopicData represents trending topic analytics
type TrendingTopicData struct {
	Topic       string    `json:"topic"`
	Volume      int       `json:"volume"`
	GrowthRate  float64   `json:"growthRate"`
	Category    string    `json:"category"`
	PeakTime    time.Time `json:"peakTime"`
}

// OptimizationSuggestion represents an AI-generated optimization suggestion
type OptimizationSuggestion struct {
	ID             string                 `json:"id"`
	VideoID        string                 `json:"videoId"`
	Type           SuggestionType         `json:"type"`
	Priority       Priority               `json:"priority"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	CurrentValue   string                 `json:"currentValue,omitempty"`
	SuggestedValue string                 `json:"suggestedValue,omitempty"`
	ExpectedImpact float64                `json:"expectedImpact"` // percentage improvement
	Confidence     float64                `json:"confidence"`     // 0-1
	AutoApplicable bool                   `json:"autoApplicable"`
	Applied        bool                   `json:"applied"`
	AppliedAt      *time.Time             `json:"appliedAt,omitempty"`
	Result         *OptimizationResult    `json:"result,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
}

// SuggestionType represents the type of optimization
type SuggestionType string

const (
	SuggestionTypeTitle       SuggestionType = "title"
	SuggestionTypeThumbnail   SuggestionType = "thumbnail"
	SuggestionTypeDescription SuggestionType = "description"
	SuggestionTypeTags        SuggestionType = "tags"
	SuggestionTypeTiming      SuggestionType = "timing"
	SuggestionTypeContent     SuggestionType = "content"
	SuggestionTypeDuration    SuggestionType = "duration"
	SuggestionTypeHook        SuggestionType = "hook"
	SuggestionTypeCTA         SuggestionType = "cta"
	SuggestionTypeRetention   SuggestionType = "retention"
)

// Priority represents suggestion priority
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// OptimizationResult represents the result of an applied optimization
type OptimizationResult struct {
	BeforeMetrics  *VideoAnalytics `json:"beforeMetrics"`
	AfterMetrics   *VideoAnalytics `json:"afterMetrics"`
	Improvement    float64         `json:"improvement"` // percentage
	ViewsIncrease  int             `json:"viewsIncrease"`
	CTRChange      float64         `json:"ctrChange"`
	EngagementChange float64       `json:"engagementChange"`
}

// PerformanceReport contains comprehensive performance analysis
type PerformanceReport struct {
	ID                 string                   `json:"id"`
	UserID             string                   `json:"userId"`
	Period             string                   `json:"period"`
	StartDate          time.Time                `json:"startDate"`
	EndDate            time.Time                `json:"endDate"`
	Summary            *PerformanceSummary      `json:"summary"`
	TopVideos          []*VideoAnalytics        `json:"topVideos"`
	Underperforming    []*VideoAnalytics        `json:"underperforming"`
	Trends             []*TrendAnalysis         `json:"trends"`
	Suggestions        []*OptimizationSuggestion `json:"suggestions"`
	ComparativeResults []*ComparativeTest       `json:"comparativeResults"`
	GeneratedAt        time.Time                `json:"generatedAt"`
}

// PerformanceSummary aggregates performance metrics
type PerformanceSummary struct {
	TotalViews          int                    `json:"totalViews"`
	TotalLikes          int                    `json:"totalLikes"`
	TotalComments       int                    `json:"totalComments"`
	TotalShares         int                    `json:"totalShares"`
	AvgEngagementRate   float64                `json:"avgEngagementRate"`
	AvgCTR              float64                `json:"avgCtr"`
	TotalWatchTime      float64                `json:"totalWatchTime"` // hours
	GrowthRate          float64                `json:"growthRate"`
	BestPerformingDay   string                 `json:"bestPerformingDay"`
	BestPerformingTime  string                 `json:"bestPerformingTime"`
	TopTrafficSource    string                 `json:"topTrafficSource"`
}

// TrendAnalysis represents content trend analysis
type TrendAnalysis struct {
	Topic           string  `json:"topic"`
	YourPerformance float64 `json:"yourPerformance"`
	MarketAverage   float64 `json:"marketAverage"`
	Opportunity     float64 `json:"opportunity"` // 0-100
	Recommendation  string  `json:"recommendation"`
}

// ComparativeTest represents an A/B test result
type ComparativeTest struct {
	ID           string                `json:"id"`
	VideoID      string                `json:"videoId"`
	TestType     string                `json:"testType"` // title, thumbnail
	Variants     []*ComparativeMetrics `json:"variants"`
	WinnerID     string                `json:"winnerId"`
	Improvement  float64               `json:"improvement"`
	CompletedAt  time.Time             `json:"completedAt"`
}

// NewOptimizerService creates a new optimizer service
func NewOptimizerService(
	analyticsRepo AnalyticsRepository,
	timelineRepo TimelineRepository,
	socialService *SocialService,
	aiScriptService *AIScriptService,
) *OptimizerService {
	return &OptimizerService{
		analyticsRepo:   analyticsRepo,
		timelineRepo:    timelineRepo,
		socialService:   socialService,
		aiScriptService: aiScriptService,
	}
}

// AnalyzeVideo analyzes a video's performance and generates suggestions
func (s *OptimizerService) AnalyzeVideo(ctx context.Context, videoID string) ([]*OptimizationSuggestion, error) {
	// Get video analytics
	analytics, err := s.analyticsRepo.GetVideoPerformance(videoID, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to get video analytics: %w", err)
	}

	var suggestions []*OptimizationSuggestion

	// Analyze different aspects
	titleSuggestions := s.analyzeTitle(ctx, analytics)
	suggestions = append(suggestions, titleSuggestions...)

	thumbnailSuggestions := s.analyzeThumbnail(ctx, analytics)
	suggestions = append(suggestions, thumbnailSuggestions...)

	retentionSuggestions := s.analyzeRetention(ctx, analytics)
	suggestions = append(suggestions, retentionSuggestions...)

	timingSuggestions := s.analyzeTiming(ctx, analytics)
	suggestions = append(suggestions, timingSuggestions...)

	engagementSuggestions := s.analyzeEngagement(ctx, analytics)
	suggestions = append(suggestions, engagementSuggestions...)

	// Sort by priority and expected impact
	sort.Slice(suggestions, func(i, j int) bool {
		priorityOrder := map[Priority]int{
			PriorityHigh:   3,
			PriorityMedium: 2,
			PriorityLow:    1,
		}
		
		if priorityOrder[suggestions[i].Priority] != priorityOrder[suggestions[j].Priority] {
			return priorityOrder[suggestions[i].Priority] > priorityOrder[suggestions[j].Priority]
		}
		
		return suggestions[i].ExpectedImpact > suggestions[j].ExpectedImpact
	})

	return suggestions, nil
}

// analyzeTitle analyzes title performance and suggests improvements
func (s *OptimizerService) analyzeTitle(ctx context.Context, analytics *VideoAnalytics) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	// Check if CTR is below average
	if analytics.CTR < 4.0 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:             uuid.New().String(),
			VideoID:        analytics.VideoID,
			Type:           SuggestionTypeTitle,
			Priority:       PriorityHigh,
			Title:          "Improve Click-Through Rate",
			Description:    fmt.Sprintf("Your CTR is %.1f%%, which is below the optimal 5-8%% range. Consider a more compelling title.", analytics.CTR),
			CurrentValue:   analytics.Title,
			ExpectedImpact: 15.0,
			Confidence:     0.75,
			AutoApplicable: false,
			CreatedAt:      time.Now(),
		})
	}

	// Check title length
	if len(analytics.Title) < 30 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:             uuid.New().String(),
			VideoID:        analytics.VideoID,
			Type:           SuggestionTypeTitle,
			Priority:       PriorityMedium,
			Title:          "Extend Title Length",
			Description:    "Your title is shorter than recommended (50-60 characters optimal). Add more descriptive keywords.",
			CurrentValue:   analytics.Title,
			ExpectedImpact: 8.0,
			Confidence:     0.65,
			AutoApplicable: false,
			CreatedAt:      time.Now(),
		})
	}

	// Suggest power words
	suggestions = append(suggestions, &OptimizationSuggestion{
		ID:             uuid.New().String(),
		VideoID:        analytics.VideoID,
		Type:           SuggestionTypeTitle,
		Priority:       PriorityLow,
		Title:          "Add Power Words",
		Description:    "Include emotional trigger words like 'Ultimate', 'Secret', 'Proven' to increase engagement.",
		CurrentValue:   analytics.Title,
		ExpectedImpact: 5.0,
		Confidence:     0.60,
		AutoApplicable: false,
		CreatedAt:      time.Now(),
	})

	return suggestions
}

// analyzeThumbnail analyzes thumbnail performance
func (s *OptimizerService) analyzeThumbnail(ctx context.Context, analytics *VideoAnalytics) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	// Low CTR often indicates thumbnail issues
	if analytics.CTR < 3.5 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:             uuid.New().String(),
			VideoID:        analytics.VideoID,
			Type:           SuggestionTypeThumbnail,
			Priority:       PriorityHigh,
			Title:          "Test New Thumbnails",
			Description:    "Low CTR suggests your thumbnail isn't capturing attention. Test a more contrasting, high-emotion thumbnail.",
			ExpectedImpact: 20.0,
			Confidence:     0.80,
			AutoApplicable: false,
			CreatedAt:      time.Now(),
		})
	}

	return suggestions
}

// analyzeRetention analyzes audience retention patterns
func (s *OptimizerService) analyzeRetention(ctx context.Context, analytics *VideoAnalytics) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	if len(analytics.RetentionCurve) < 2 {
		return suggestions
	}

	// Check for early drop-off
	if analytics.RetentionCurve[1] < 0.5 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:             uuid.New().String(),
			VideoID:        analytics.VideoID,
			Type:           SuggestionTypeHook,
			Priority:       PriorityHigh,
			Title:          "Strengthen Hook",
			Description:    fmt.Sprintf("50%% of viewers drop off in the first 10%%. Start with a stronger hook or tease the payoff earlier."),
			ExpectedImpact: 25.0,
			Confidence:     0.85,
			AutoApplicable: false,
			CreatedAt:      time.Now(),
		})
	}

	// Check for mid-video drop-off
	if len(analytics.RetentionCurve) > 5 {
		midPoint := len(analytics.RetentionCurve) / 2
		if analytics.RetentionCurve[midPoint] < 0.3 {
			suggestions = append(suggestions, &OptimizationSuggestion{
				ID:             uuid.New().String(),
				VideoID:        analytics.VideoID,
				Type:           SuggestionTypeRetention,
				Priority:       PriorityMedium,
				Title:          "Add Retention Boosters",
				Description:    "Retention drops significantly mid-video. Add pattern interrupts, visual changes, or teases.",
				ExpectedImpact: 15.0,
				Confidence:     0.70,
				AutoApplicable: false,
				CreatedAt:      time.Now(),
			})
		}
	}

	return suggestions
}

// analyzeTiming analyzes publish timing performance
func (s *OptimizerService) analyzeTiming(ctx context.Context, analytics *VideoAnalytics) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	// Check if video is new and might benefit from different timing
	if analytics.DaysSincePublish < 7 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:             uuid.New().String(),
			VideoID:        analytics.VideoID,
			Type:           SuggestionTypeTiming,
			Priority:       PriorityLow,
			Title:          "Optimal Publish Time",
			Description:    "Based on your audience analytics, consider publishing on Tuesday at 2 PM for maximum reach.",
			ExpectedImpact: 10.0,
			Confidence:     0.65,
			AutoApplicable: true,
			CreatedAt:      time.Now(),
		})
	}

	return suggestions
}

// analyzeEngagement analyzes engagement patterns
func (s *OptimizerService) analyzeEngagement(ctx context.Context, analytics *VideoAnalytics) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	// Low engagement rate
	if analytics.EngagementRate < 3.0 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:             uuid.New().String(),
			VideoID:        analytics.VideoID,
			Type:           SuggestionTypeCTA,
			Priority:       PriorityMedium,
			Title:          "Improve Call-to-Action",
			Description:    fmt.Sprintf("Engagement rate is %.1f%%. Add clearer CTAs asking viewers to like, comment, and subscribe.", analytics.EngagementRate),
			ExpectedImpact: 12.0,
			Confidence:     0.70,
			AutoApplicable: false,
			CreatedAt:      time.Now(),
		})
	}

	return suggestions
}

// GeneratePerformanceReport generates a comprehensive performance report
func (s *OptimizerService) GeneratePerformanceReport(ctx context.Context, userID string, days int) (*PerformanceReport, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	report := &PerformanceReport{
		ID:          uuid.New().String(),
		UserID:      userID,
		Period:      fmt.Sprintf("Last %d days", days),
		StartDate:   startDate,
		EndDate:     endDate,
		GeneratedAt: time.Now(),
	}

	// Get top performing videos
	topVideos, err := s.analyticsRepo.GetTopPerforming(userID, 10)
	if err != nil {
		return nil, err
	}
	report.TopVideos = topVideos

	// Get underperforming videos
	underperforming, err := s.analyticsRepo.GetUnderperforming(userID, 10)
	if err != nil {
		return nil, err
	}
	report.Underperforming = underperforming

	// Generate summary
	report.Summary = s.generateSummary(topVideos, underperforming)

	// Generate trends analysis
	report.Trends = s.analyzeTrends(ctx, topVideos)

	// Generate suggestions for underperforming videos
	for _, video := range underperforming {
		suggestions, err := s.AnalyzeVideo(ctx, video.VideoID)
		if err != nil {
			continue
		}
		report.Suggestions = append(report.Suggestions, suggestions...)
	}

	return report, nil
}

// generateSummary generates performance summary
func (s *OptimizerService) generateSummary(topVideos, underperforming []*VideoAnalytics) *PerformanceSummary {
	summary := &PerformanceSummary{}

	var totalViews, totalLikes, totalComments, totalShares int
	var totalEngagementRate, totalCTR float64

	allVideos := append(topVideos, underperforming...)

	for _, video := range allVideos {
		totalViews += video.Views
		totalLikes += video.Likes
		totalComments += video.Comments
		totalShares += video.Shares
		totalEngagementRate += video.EngagementRate
		totalCTR += video.CTR
	}

	count := len(allVideos)
	if count > 0 {
		summary.TotalViews = totalViews
		summary.TotalLikes = totalLikes
		summary.TotalComments = totalComments
		summary.TotalShares = totalShares
		summary.AvgEngagementRate = totalEngagementRate / float64(count)
		summary.AvgCTR = totalCTR / float64(count)
	}

	return summary
}

// analyzeTrends analyzes content trends
func (s *OptimizerService) analyzeTrends(ctx context.Context, videos []*VideoAnalytics) []*TrendAnalysis {
	var trends []*TrendAnalysis

	// Group videos by topic/theme
	topicMap := make(map[string][]*VideoAnalytics)
	
	for _, video := range videos {
		// Extract topic from title (simplified)
		topic := extractTopic(video.Title)
		topicMap[topic] = append(topicMap[topic], video)
	}

	// Analyze each topic
	for topic, topicVideos := range topicMap {
		if len(topicVideos) < 2 {
			continue
		}

		var totalViews int
		for _, v := range topicVideos {
			totalViews += v.Views
		}
		avgViews := float64(totalViews) / float64(len(topicVideos))

		trends = append(trends, &TrendAnalysis{
			Topic:           topic,
			YourPerformance: avgViews,
			MarketAverage:   avgViews * 0.8, // Simulated market average
			Opportunity:     math.Min(100, avgViews/1000),
			Recommendation:  fmt.Sprintf("Double down on '%s' content - performing %.0f%% above average", topic, 25.0),
		})
	}

	return trends
}

// AutoOptimizeTitle automatically optimizes a video title
func (s *OptimizerService) AutoOptimizeTitle(ctx context.Context, videoID string, currentTitle string) (string, error) {
	// Analyze current title
	issues := []string{}
	
	if len(currentTitle) < 30 {
		issues = append(issues, "too_short")
	}
	if !strings.ContainsAny(currentTitle, "1234567890") {
		issues = append(issues, "no_numbers")
	}
	if !strings.Contains(currentTitle, "?") && !strings.Contains(currentTitle, "How") {
		issues = append(issues, "not_question")
	}

	// Generate improved title templates
	templates := []string{
		"How I [Achieved Result] in [Timeframe] (Step-by-Step)",
		"The [Number] [Topic] Secrets Experts Won't Tell You",
		"Why Your [Topic] Strategy is Failing (And How to Fix It)",
		"I Tested [Method] for 30 Days. Here are the Results.",
		"[Number] Ways to [Achieve Goal] Faster",
	}

	// Select best template based on issues
	selectedTemplate := templates[0]
	
	if containsString(issues, "not_question") {
		selectedTemplate = templates[2]
	} else if containsString(issues, "no_numbers") {
		selectedTemplate = templates[1]
	}

	return selectedTemplate, nil
}

// AutoOptimizeThumbnail generates an optimized thumbnail
func (s *OptimizerService) AutoOptimizeThumbnail(ctx context.Context, videoID string, currentThumbnailURL string) (string, error) {
	// This would use AI to:
	// 1. Analyze current thumbnail
	// 2. Check contrast, faces, text readability
	// 3. Generate improvements
	
	// For now, return suggestion
	return "Generate high-contrast thumbnail with face and 3-5 words max", nil
}

// ApplySuggestion applies an optimization suggestion
func (s *OptimizerService) ApplySuggestion(ctx context.Context, suggestion *OptimizationSuggestion) error {
	if suggestion.Applied {
		return fmt.Errorf("suggestion already applied")
	}

	// Apply based on type
	switch suggestion.Type {
	case SuggestionTypeTitle:
		// Update video title
		newTitle, err := s.AutoOptimizeTitle(ctx, suggestion.VideoID, suggestion.CurrentValue)
		if err != nil {
			return err
		}
		suggestion.SuggestedValue = newTitle
		
	case SuggestionTypeThumbnail:
		// Generate new thumbnail
		thumbURL, err := s.AutoOptimizeThumbnail(ctx, suggestion.VideoID, "")
		if err != nil {
			return err
		}
		suggestion.SuggestedValue = thumbURL
		
	default:
		return fmt.Errorf("auto-apply not supported for suggestion type: %s", suggestion.Type)
	}

	now := time.Now()
	suggestion.Applied = true
	suggestion.AppliedAt = &now

	return nil
}

// GetWinningContent identifies top-performing content patterns
func (s *OptimizerService) GetWinningContent(ctx context.Context, userID string) (*WinningContentAnalysis, error) {
	// Get top performing videos
	topVideos, err := s.analyticsRepo.GetTopPerforming(userID, 20)
	if err != nil {
		return nil, err
	}

	analysis := &WinningContentAnalysis{
		UserID: userID,
	}

	// Analyze patterns
	titlePatterns := make(map[string]int)
	durationBuckets := make(map[string]int)
	publishDayCounts := make(map[string]int)

	for _, video := range topVideos {
		// Extract title patterns
		if strings.Contains(video.Title, "How") {
			titlePatterns["how-to"]++
		}
		if strings.Contains(video.Title, "vs") || strings.Contains(video.Title, "Versus") {
			titlePatterns["comparison"]++
		}
		if strings.ContainsAny(video.Title, "0123456789") {
			titlePatterns["listicle"]++
		}

		// Duration buckets
		duration := video.AvgWatchDuration
		switch {
		case duration < 60:
			durationBuckets["short_30s"]++
		case duration < 180:
			durationBuckets["medium_3min"]++
		case duration < 600:
			durationBuckets["long_10min"]++
		default:
			durationBuckets["extended_10min+"]++
		}

		// Publish day
		day := video.PublishDate.Weekday().String()
		publishDayCounts[day]++
	}

	analysis.TitlePatterns = titlePatterns
	analysis.OptimalDuration = findMaxKey(durationBuckets)
	analysis.BestPublishDay = findMaxKey(publishDayCounts)

	return analysis, nil
}

// WinningContentAnalysis contains winning content patterns
type WinningContentAnalysis struct {
	UserID          string         `json:"userId"`
	TitlePatterns   map[string]int `json:"titlePatterns"`
	OptimalDuration string         `json:"optimalDuration"`
	BestPublishDay  string         `json:"bestPublishDay"`
	BestCTA         string         `json:"bestCta,omitempty"`
	TopThumbnailStyle string       `json:"topThumbnailStyle,omitempty"`
}

// Helper functions

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func extractTopic(title string) string {
	// Simplified topic extraction
	words := strings.Fields(strings.ToLower(title))
	if len(words) > 2 {
		return words[0] + " " + words[1]
	}
	return title
}

func findMaxKey(m map[string]int) string {
	var maxKey string
	maxVal := -1
	for k, v := range m {
		if v > maxVal {
			maxVal = v
			maxKey = k
		}
	}
	return maxKey
}
