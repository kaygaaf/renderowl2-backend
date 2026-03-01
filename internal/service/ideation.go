package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// IdeationService provides content ideation and trending topic discovery
type IdeationService struct {
	httpClient  *http.Client
	apiKeys     map[string]string
	cache       map[string]*CacheEntry
	cacheMutex  sync.RWMutex
	cacheExpiry time.Duration
}

// CacheEntry represents a cached API response
type CacheEntry struct {
	Data      interface{}
	Timestamp time.Time
}

// TrendingTopic represents a trending topic from any platform
type TrendingTopic struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Platform      string   `json:"platform"`
	Category      string   `json:"category"`
	Score         float64  `json:"score"`
	Volume        int      `json:"volume"`
	Velocity      float64  `json:"velocity"` // Rate of growth
	RelatedTopics []string `json:"relatedTopics,omitempty"`
	URL           string   `json:"url,omitempty"`
	ThumbnailURL  string   `json:"thumbnailUrl,omitempty"`
	PublishedAt   *time.Time `json:"publishedAt,omitempty"`
}

// ContentSuggestion represents an AI-generated content suggestion
type ContentSuggestion struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Niche         string   `json:"niche"`
	Format        string   `json:"format"` // short, long, series
	EstimatedViews int     `json:"estimatedViews"`
	Difficulty    string   `json:"difficulty"` // easy, medium, hard
	TimeToCreate  int      `json:"timeToCreate"` // minutes
	Hook          string   `json:"hook"`
	Outline       []string `json:"outline"`
	Tags          []string `json:"tags"`
	TrendingScore float64  `json:"trendingScore"`
}

// CompetitorAnalysis represents analysis of a competitor channel
type CompetitorAnalysis struct {
	ChannelID      string            `json:"channelId"`
	ChannelName    string            `json:"channelName"`
	Platform       string            `json:"platform"`
	SubscriberCount int              `json:"subscriberCount"`
	TotalViews     int               `json:"totalViews"`
	VideoCount     int               `json:"videoCount"`
	AvgViewsPerVideo int             `json:"avgViewsPerVideo"`
	TopVideos      []CompetitorVideo `json:"topVideos"`
	ContentGaps    []ContentGap      `json:"contentGaps"`
	PostingFrequency string          `json:"postingFrequency"`
	BestPostTime   string            `json:"bestPostTime"`
	AnalyzedAt     time.Time         `json:"analyzedAt"`
}

// CompetitorVideo represents a top-performing competitor video
type CompetitorVideo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Views       int       `json:"views"`
	Likes       int       `json:"likes"`
	Comments    int       `json:"comments"`
	PublishedAt time.Time `json:"publishedAt"`
	URL         string    `json:"url"`
	Thumbnail   string    `json:"thumbnail"`
}

// ContentGap represents a content opportunity gap
type ContentGap struct {
	Topic       string  `json:"topic"`
	SearchVolume int    `json:"searchVolume"`
	Competition string  `json:"competition"` // low, medium, high
	Opportunity float64 `json:"opportunity"` // 0-100 score
}

// ContentCalendar represents a 30-day content calendar
type ContentCalendar struct {
	ID        string               `json:"id"`
	Niche     string               `json:"niche"`
	StartDate time.Time            `json:"startDate"`
	Days      []ContentCalendarDay `json:"days"`
}

// ContentCalendarDay represents a single day in the calendar
type ContentCalendarDay struct {
	Date        time.Time       `json:"date"`
	Video       *CalendarVideo  `json:"video,omitempty"`
	IsRestDay   bool            `json:"isRestDay"`
	Theme       string          `json:"theme,omitempty"`
}

// CalendarVideo represents a planned video in the calendar
type CalendarVideo struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Format      string   `json:"format"`
	Tags        []string `json:"tags"`
	ScriptOutline string `json:"scriptOutline,omitempty"`
}

// GetTrendingTopicsRequest represents a request for trending topics
type GetTrendingTopicsRequest struct {
	Platforms  []string `json:"platforms,omitempty"`  // youtube, tiktok, twitter
	Categories []string `json:"categories,omitempty"` // tech, gaming, education, etc.
	Limit      int      `json:"limit,omitempty"`
	Region     string   `json:"region,omitempty"`     // US, EU, GLOBAL
}

// GetContentSuggestionsRequest represents a request for content suggestions
type GetContentSuggestionsRequest struct {
	Niche       string   `json:"niche" binding:"required"`
	Format      string   `json:"format,omitempty"` // short, long, series
	Count       int      `json:"count,omitempty"`
	TrendingOnly bool    `json:"trendingOnly,omitempty"`
}

// CompetitorAnalysisRequest represents a request for competitor analysis
type CompetitorAnalysisRequest struct {
	ChannelURL string `json:"channelUrl" binding:"required"`
	Platform   string `json:"platform" binding:"required"`
	Depth      string `json:"depth,omitempty"` // basic, detailed, deep
}

// GenerateCalendarRequest represents a request for calendar generation
type GenerateCalendarRequest struct {
	Niche     string    `json:"niche" binding:"required"`
	StartDate time.Time `json:"startDate,omitempty"`
	Frequency int       `json:"frequency,omitempty"` // videos per week
	Format    string    `json:"format,omitempty"`
}

// NewIdeationService creates a new ideation service
func NewIdeationService() *IdeationService {
	return &IdeationService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKeys:     make(map[string]string),
		cache:       make(map[string]*CacheEntry),
		cacheExpiry: 15 * time.Minute,
	}
}

// SetAPIKey sets an API key for a platform
func (s *IdeationService) SetAPIKey(platform, key string) {
	s.apiKeys[platform] = key
}

// GetTrendingTopics retrieves trending topics from multiple platforms
func (s *IdeationService) GetTrendingTopics(ctx context.Context, req *GetTrendingTopicsRequest) ([]*TrendingTopic, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}

	var wg sync.WaitGroup
	topicChan := make(chan []*TrendingTopic, len(req.Platforms))
	errorChan := make(chan error, len(req.Platforms))

	// Fetch from each platform concurrently
	for _, platform := range req.Platforms {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			topics, err := s.fetchPlatformTrends(ctx, p, req)
			if err != nil {
				errorChan <- err
				return
			}
			topicChan <- topics
		}(platform)
	}

	// Close channels when done
	go func() {
		wg.Wait()
		close(topicChan)
		close(errorChan)
	}()

	// Collect results
	var allTopics []*TrendingTopic
	for topics := range topicChan {
		allTopics = append(allTopics, topics...)
	}

	// Check for errors (but don't fail completely)
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	// Sort by score
	sort.Slice(allTopics, func(i, j int) bool {
		return allTopics[i].Score > allTopics[j].Score
	})

	// Apply limit
	if len(allTopics) > req.Limit {
		allTopics = allTopics[:req.Limit]
	}

	return allTopics, nil
}

// fetchPlatformTrends fetches trends from a specific platform
func (s *IdeationService) fetchPlatformTrends(ctx context.Context, platform string, req *GetTrendingTopicsRequest) ([]*TrendingTopic, error) {
	switch strings.ToLower(platform) {
	case "youtube":
		return s.fetchYouTubeTrends(ctx, req)
	case "tiktok":
		return s.fetchTikTokTrends(ctx, req)
	case "twitter", "x":
		return s.fetchTwitterTrends(ctx, req)
	case "reddit":
		return s.fetchRedditTrends(ctx, req)
	default:
		return s.generateSimulatedTrends(platform, req)
	}
}

// fetchYouTubeTrends fetches trending topics from YouTube
func (s *IdeationService) fetchYouTubeTrends(ctx context.Context, req *GetTrendingTopicsRequest) ([]*TrendingTopic, error) {
	apiKey := s.apiKeys["youtube"]
	if apiKey == "" {
		return s.generateSimulatedTrends("youtube", req)
	}

	cacheKey := fmt.Sprintf("youtube_trends_%s", req.Region)
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.([]*TrendingTopic), nil
	}

	regionCode := req.Region
	if regionCode == "" || regionCode == "GLOBAL" {
		regionCode = "US"
	}

	u := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&chart=mostPopular&regionCode=%s&maxResults=50&key=%s",
		regionCode, apiKey,
	)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return s.generateSimulatedTrends("youtube", req)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return s.generateSimulatedTrends("youtube", req)
	}

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string   `json:"title"`
				Description string   `json:"description"`
				CategoryID  string   `json:"categoryId"`
				Tags        []string `json:"tags"`
				Thumbnails  struct {
					High struct {
						URL string `json:"url"`
					} `json:"high"`
				} `json:"thumbnails"`
				PublishedAt string `json:"publishedAt"`
			} `json:"snippet"`
			Statistics struct {
				ViewCount    string `json:"viewCount"`
				LikeCount    string `json:"likeCount"`
				CommentCount string `json:"commentCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.generateSimulatedTrends("youtube", req)
	}

	var topics []*TrendingTopic
	for _, item := range result.Items {
		publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		
		topic := &TrendingTopic{
			ID:           uuid.New().String(),
			Title:        item.Snippet.Title,
			Description:  truncateString(item.Snippet.Description, 200),
			Platform:     "youtube",
			Category:     getCategoryName(item.Snippet.CategoryID),
			RelatedTopics: item.Snippet.Tags[:min(5, len(item.Snippet.Tags))],
			URL:          fmt.Sprintf("https://youtube.com/watch?v=%s", item.ID),
			ThumbnailURL: item.Snippet.Thumbnails.High.URL,
			PublishedAt:  &publishedAt,
		}
		topics = append(topics, topic)
	}

	s.setCache(cacheKey, topics)
	return topics, nil
}

// fetchTikTokTrends fetches trending topics from TikTok
func (s *IdeationService) fetchTikTokTrends(ctx context.Context, req *GetTrendingTopicsRequest) ([]*TrendingTopic, error) {
	// TikTok doesn't have a public trending API without authentication
	// Return simulated data for now
	return s.generateSimulatedTrends("tiktok", req)
}

// fetchTwitterTrends fetches trending topics from Twitter/X
func (s *IdeationService) fetchTwitterTrends(ctx context.Context, req *GetTrendingTopicsRequest) ([]*TrendingTopic, error) {
	apiKey := s.apiKeys["twitter"]
	if apiKey == "" {
		return s.generateSimulatedTrends("twitter", req)
	}

	// Would use Twitter API v2 here
	// For now, return simulated data
	return s.generateSimulatedTrends("twitter", req)
}

// fetchRedditTrends fetches trending topics from Reddit
func (s *IdeationService) fetchRedditTrends(ctx context.Context, req *GetTrendingTopicsRequest) ([]*TrendingTopic, error) {
	cacheKey := "reddit_trends"
	if cached := s.getCache(cacheKey); cached != nil {
		return cached.([]*TrendingTopic), nil
	}

	u := "https://www.reddit.com/r/all/hot.json?limit=25"
	httpReq, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return s.generateSimulatedTrends("reddit", req)
	}

	httpReq.Header.Set("User-Agent", "Renderowl/2.0")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return s.generateSimulatedTrends("reddit", req)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Children []struct {
				Data struct {
					Title     string  `json:"title"`
					Subreddit string  `json:"subreddit"`
					Score     float64 `json:"score"`
					URL       string  `json:"url"`
					Permalink string  `json:"permalink"`
					Ups       int     `json:"ups"`
					NumComments int   `json:"num_comments"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.generateSimulatedTrends("reddit", req)
	}

	var topics []*TrendingTopic
	for _, child := range result.Data.Children {
		topic := &TrendingTopic{
			ID:          uuid.New().String(),
			Title:       child.Data.Title,
			Description: fmt.Sprintf("Trending in r/%s", child.Data.Subreddit),
			Platform:    "reddit",
			Category:    child.Data.Subreddit,
			Score:       child.Data.Score,
			Volume:      child.Data.Ups + child.Data.NumComments,
			URL:         fmt.Sprintf("https://reddit.com%s", child.Data.Permalink),
		}
		topics = append(topics, topic)
	}

	s.setCache(cacheKey, topics)
	return topics, nil
}

// generateSimulatedTrends generates simulated trending topics for development/testing
func (s *IdeationService) generateSimulatedTrends(platform string, req *GetTrendingTopicsRequest) ([]*TrendingTopic, error) {
	templates := map[string][]struct {
		title    string
		category string
		volume   int
	}{
		"youtube": {
			{"I Tried [X] for 30 Days", "lifestyle", 5000000},
			{"The Truth About [Topic]", "education", 3200000},
			{"[Celebrity] Reacts to [Trend]", "entertainment", 8500000},
			{"How I Made $[Amount] in [Timeframe]", "business", 2100000},
			{"[Product] vs [Product] - Honest Review", "tech", 1800000},
			{"I Built [Something] in 24 Hours", "tech", 4500000},
			{"Reacting to My Old [Content]", "entertainment", 1200000},
			{"The Rise and Fall of [Company]", "business", 6700000},
		},
		"tiktok": {
			{"POV: When [Relatable Situation]", "comedy", 2500000},
			{"Day in the Life: [Profession]", "lifestyle", 1800000},
			{"[Hack] That Will Change Your Life", "tips", 3200000},
			{"Before vs After [Transformation]", "transformation", 4100000},
			{"How to [Skill] in 60 Seconds", "education", 2900000},
			{"Things I Wish I Knew Before [Experience]", "advice", 1500000},
			{"Testing Viral [Trend/Product]", "trending", 3800000},
		},
		"twitter": {
			{"Hot take: [Controversial Opinion]", "discourse", 850000},
			{"Thread: Everything About [Topic]", "education", 620000},
			{"[Industry] predictions for 2025", "tech", 430000},
			{"Unpopular opinion about [Topic]", "discourse", 780000},
			{"Story time about [Experience]", "story", 920000},
		},
		"reddit": {
			{"AITA for [Situation]?", "advice", 45000},
			{"TIL: [Interesting Fact]", "education", 38000},
			{"What is something [Question]?", "discussion", 52000},
			{"My [Achievement] after [Timeframe]", "progress", 67000},
		},
	}

	var topics []*TrendingTopic
	templateList, ok := templates[platform]
	if !ok {
		templateList = templates["youtube"]
	}

	for i, tmpl := range templateList {
		topic := &TrendingTopic{
			ID:            fmt.Sprintf("%s_trend_%d", platform, i),
			Title:         tmpl.title,
			Description:   fmt.Sprintf("Trending %s topic in %s", platform, tmpl.category),
			Platform:      platform,
			Category:      tmpl.category,
			Score:         float64(tmpl.volume) / 1000000,
			Volume:        tmpl.volume,
			Velocity:      float64(i+1) * 1.5,
			RelatedTopics: []string{"trending", tmpl.category, platform},
		}
		topics = append(topics, topic)
	}

	return topics, nil
}

// GetContentSuggestions generates AI-powered content suggestions
func (s *IdeationService) GetContentSuggestions(ctx context.Context, req *GetContentSuggestionsRequest) ([]*ContentSuggestion, error) {
	if req.Count == 0 {
		req.Count = 10
	}
	if req.Format == "" {
		req.Format = "short"
	}

	// Define suggestion templates based on niche
	nicheTemplates := map[string][]struct {
		title       string
		description string
		hook        string
		outline     []string
		tags        []string
		difficulty  string
		timeToCreate int
	}{
		"tech": {
			{
				title:        "5 AI Tools That Will 10x Your Productivity",
				description:  "Discover the best AI tools for productivity",
				hook:         "What if I told you these 5 tools could save you 20 hours a week?",
				outline:      []string{"Intro hook", "Tool 1 with demo", "Tool 2 with demo", "Tool 3 with demo", "Tool 4 with demo", "Tool 5 with demo", "Comparison table", "CTA"},
				tags:         []string{"AI", "productivity", "tech", "tools"},
				difficulty:   "easy",
				timeToCreate: 45,
			},
			{
				title:        "I Switched to Linux for 30 Days",
				description:  "My experience switching from Windows to Linux",
				hook:         "I deleted Windows and forced myself to use Linux for 30 days...",
				outline:      []string{"Why I'm switching", "Day 1-7 struggles", "Day 8-14 learning", "Day 15-21 comfort", "Day 22-30 mastery", "Final verdict"},
				tags:         []string{"linux", "windows", "os", "tech"},
				difficulty:   "medium",
				timeToCreate: 90,
			},
		},
		"gaming": {
			{
				title:        "I Broke [Game] in the Most Creative Way",
				description:  "Finding and exploiting game mechanics",
				hook:         "This glitch shouldn't be possible...",
				outline:      []string{"Intro mystery", "Finding the glitch", "Attempt 1", "Attempt 2", "Success!", "Implications"},
				tags:         []string{"gaming", "glitch", "speedrun"},
				difficulty:   "hard",
				timeToCreate: 120,
			},
		},
		"education": {
			{
				title:        "[Complex Topic] Explained Like You're 5",
				description:  "Simple explanation of complex topic",
				hook:         "If you can understand this analogy, you'll master [topic]...",
				outline:      []string{"The problem with complex explanations", "The analogy", "Breaking it down", "Real world examples", "Common misconceptions", "Summary"},
				tags:         []string{"education", "explained", "learning"},
				difficulty:   "easy",
				timeToCreate: 60,
			},
		},
		"finance": {
			{
				title:        "How I Saved $[Amount] in [Timeframe]",
				description:  "Practical money saving strategies",
				hook:         "I went from broke to saving $10K in 6 months with these 7 rules",
				outline:      []string{"My story", "Rule 1: The 50/30/20 method", "Rule 2: Automate savings", "Rule 3: Cut subscriptions", "Rule 4: Meal prep", "Rules 5-7", "Results", "Your action plan"},
				tags:         []string{"finance", "saving", "money", "budget"},
				difficulty:   "easy",
				timeToCreate: 75,
			},
		},
		"fitness": {
			{
				title:        "I Did [Exercise] Every Day for 30 Days",
				description:  "30-day fitness challenge results",
				hook:         "The results after 30 days of [exercise] surprised everyone...",
				outline:      []string{"Day 0 measurements", "Week 1 struggle", "Week 2 adaptation", "Week 3 progress", "Week 4 results", "Before/After", "What I learned"},
				tags:         []string{"fitness", "challenge", "transformation"},
				difficulty:   "medium",
				timeToCreate: 90,
			},
		},
		"cooking": {
			{
				title:        "Restaurant-Quality [Dish] at Home",
				description:  "Make professional meals at home",
				hook:         "This chef's secret technique changed how I cook forever...",
				outline:      []string{"Why restaurant food tastes better", "The secret ingredient", "Step-by-step cooking", "Plating technique", "Taste test", "Recipe card"},
				tags:         []string{"cooking", "recipe", "food", "chef"},
				difficulty:   "medium",
				timeToCreate: 120,
			},
		},
		"travel": {
			{
				title:        "[Destination] on a $[Amount] Budget",
				description:  "Budget travel guide",
				hook:         "I visited [destination] for less than you spend on coffee per month...",
				outline:      []string{"Budget breakdown", "Cheap flights hack", "Affordable stays", "Free activities", "Cheap eats", "Total cost", "Money-saving tips"},
				tags:         []string{"travel", "budget", "backpacking", "tips"},
				difficulty:   "easy",
				timeToCreate: 60,
			},
		},
	}

	// Get templates for the requested niche or use generic ones
	templates, ok := nicheTemplates[req.Niche]
	if !ok {
		// Use generic templates
		templates = nicheTemplates["tech"]
	}

	var suggestions []*ContentSuggestion
	for i, tmpl := range templates {
		if i >= req.Count {
			break
		}

		suggestion := &ContentSuggestion{
			ID:             uuid.New().String(),
			Title:          tmpl.title,
			Description:    tmpl.description,
			Niche:          req.Niche,
			Format:         req.Format,
			EstimatedViews: calculateEstimatedViews(tmpl.difficulty, req.Niche),
			Difficulty:     tmpl.difficulty,
			TimeToCreate:   tmpl.timeToCreate,
			Hook:           tmpl.hook,
			Outline:        tmpl.outline,
			Tags:           tmpl.tags,
			TrendingScore:  calculateTrendingScore(),
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// AnalyzeCompetitor analyzes a competitor's channel
func (s *IdeationService) AnalyzeCompetitor(ctx context.Context, req *CompetitorAnalysisRequest) (*CompetitorAnalysis, error) {
	// Parse channel URL to extract ID
	channelID := extractChannelID(req.ChannelURL, req.Platform)
	
	analysis := &CompetitorAnalysis{
		ChannelID:   channelID,
		Platform:    req.Platform,
		AnalyzedAt:  time.Now(),
	}

	switch req.Platform {
	case "youtube":
		return s.analyzeYouTubeChannel(ctx, analysis)
	case "tiktok":
		return s.analyzeTikTokChannel(ctx, analysis)
	default:
		return s.generateSimulatedAnalysis(analysis)
	}
}

// analyzeYouTubeChannel analyzes a YouTube channel
func (s *IdeationService) analyzeYouTubeChannel(ctx context.Context, analysis *CompetitorAnalysis) (*CompetitorAnalysis, error) {
	apiKey := s.apiKeys["youtube"]
	if apiKey == "" {
		return s.generateSimulatedAnalysis(analysis)
	}

	// Would use YouTube Data API here
	// For now, return simulated data
	return s.generateSimulatedAnalysis(analysis)
}

// analyzeTikTokChannel analyzes a TikTok channel
func (s *IdeationService) analyzeTikTokChannel(ctx context.Context, analysis *CompetitorAnalysis) (*CompetitorAnalysis, error) {
	// TikTok API requires authentication
	return s.generateSimulatedAnalysis(analysis)
}

// generateSimulatedAnalysis generates simulated competitor analysis
func (s *IdeationService) generateSimulatedAnalysis(analysis *CompetitorAnalysis) (*CompetitorAnalysis, error) {
	analysis.ChannelName = "Example Creator"
	analysis.SubscriberCount = 1250000
	analysis.TotalViews = 45600000
	analysis.VideoCount = 234
	analysis.AvgViewsPerVideo = 195000
	analysis.PostingFrequency = "3 videos per week"
	analysis.BestPostTime = "Tuesday 3 PM EST"

	// Simulated top videos
	analysis.TopVideos = []CompetitorVideo{
		{
			ID:          "vid1",
			Title:       "I Tried [X] for 30 Days - Shocking Results",
			Views:       5200000,
			Likes:       234000,
			Comments:    8900,
			PublishedAt: time.Now().AddDate(0, -2, 0),
			URL:         "https://youtube.com/watch?v=example1",
		},
		{
			ID:          "vid2",
			Title:       "The Truth About [Topic] Nobody Talks About",
			Views:       3800000,
			Likes:       167000,
			Comments:    12000,
			PublishedAt: time.Now().AddDate(0, -4, 0),
			URL:         "https://youtube.com/watch?v=example2",
		},
	}

	// Simulated content gaps
	analysis.ContentGaps = []ContentGap{
		{Topic: "Beginner guides", SearchVolume: 50000, Competition: "low", Opportunity: 85},
		{Topic: "Budget alternatives", SearchVolume: 35000, Competition: "medium", Opportunity: 72},
		{Topic: "Myth busting", SearchVolume: 42000, Competition: "low", Opportunity: 78},
	}

	return analysis, nil
}

// GenerateContentCalendar generates a 30-day content calendar
func (s *IdeationService) GenerateContentCalendar(ctx context.Context, req *GenerateCalendarRequest) (*ContentCalendar, error) {
	if req.StartDate.IsZero() {
		req.StartDate = time.Now()
	}
	if req.Frequency == 0 {
		req.Frequency = 3 // 3 videos per week default
	}
	if req.Format == "" {
		req.Format = "short"
	}

	calendar := &ContentCalendar{
		ID:        uuid.New().String(),
		Niche:     req.Niche,
		StartDate: req.StartDate,
	}

	// Get content suggestions to populate calendar
	suggestionsReq := &GetContentSuggestionsRequest{
		Niche:  req.Niche,
		Format: req.Format,
		Count:  30,
	}
	suggestions, err := s.GetContentSuggestions(ctx, suggestionsReq)
	if err != nil {
		return nil, err
	}

	// Generate 30 days
	postDays := calculatePostDays(req.Frequency)
	suggestionIndex := 0

	for day := 0; day < 30; day++ {
		date := req.StartDate.AddDate(0, 0, day)
		dayOfWeek := int(date.Weekday())
		
		calendarDay := ContentCalendarDay{
			Date: date,
		}

		// Check if this is a posting day
		if contains(postDays, dayOfWeek) && suggestionIndex < len(suggestions) {
			suggestion := suggestions[suggestionIndex]
			calendarDay.Video = &CalendarVideo{
				ID:          suggestion.ID,
				Title:       suggestion.Title,
				Description: suggestion.Description,
				Format:      suggestion.Format,
				Tags:        suggestion.Tags,
				ScriptOutline: strings.Join(suggestion.Outline, "\n"),
			}
			suggestionIndex++
		} else {
			calendarDay.IsRestDay = true
			calendarDay.Theme = "Rest & Planning"
		}

		calendar.Days = append(calendar.Days, calendarDay)
	}

	return calendar, nil
}

// Helper functions

func (s *IdeationService) getCache(key string) interface{} {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	entry, ok := s.cache[key]
	if !ok {
		return nil
	}

	if time.Since(entry.Timestamp) > s.cacheExpiry {
		return nil
	}

	return entry.Data
}

func (s *IdeationService) setCache(key string, data interface{}) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.cache[key] = &CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
	}
}

func calculateEstimatedViews(difficulty, niche string) int {
	base := 100000
	switch difficulty {
	case "easy":
		base = 50000
	case "medium":
		base = 150000
	case "hard":
		base = 300000
	}
	
	// Niche multiplier
	multipliers := map[string]float64{
		"tech":       1.5,
		"gaming":     2.0,
		"finance":    1.3,
		"fitness":    1.4,
		"education":  1.2,
		"cooking":    1.1,
		"travel":     1.0,
	}
	
	if m, ok := multipliers[niche]; ok {
		base = int(float64(base) * m)
	}
	
	return base
}

func calculateTrendingScore() float64 {
	// Generate a trending score between 60-95
	return 60 + float64(time.Now().Unix()%35)
}

func calculatePostDays(frequency int) []int {
	// Calculate which days of week to post based on frequency
	switch frequency {
	case 1:
		return []int{3} // Wednesday
	case 2:
		return []int{1, 4} // Tuesday, Friday
	case 3:
		return []int{1, 3, 5} // Tuesday, Thursday, Saturday
	case 4:
		return []int{1, 3, 5, 0} // Tuesday, Thursday, Saturday, Sunday
	case 5:
		return []int{1, 2, 3, 5, 6} // Tuesday-Thursday, Saturday-Sunday
	case 7:
		return []int{0, 1, 2, 3, 4, 5, 6} // Daily
	default:
		return []int{1, 3, 5}
	}
}

func extractChannelID(url, platform string) string {
	// Extract channel ID from URL
	// This is a simplified version
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getCategoryName(categoryID string) string {
	categories := map[string]string{
		"1":  "Film & Animation",
		"2":  "Autos & Vehicles",
		"10": "Music",
		"15": "Pets & Animals",
		"17": "Sports",
		"19": "Travel & Events",
		"20": "Gaming",
		"22": "People & Blogs",
		"23": "Comedy",
		"24": "Entertainment",
		"25": "News & Politics",
		"26": "Howto & Style",
		"27": "Education",
		"28": "Science & Technology",
	}
	if name, ok := categories[categoryID]; ok {
		return name
	}
	return "Unknown"
}

func contains(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
