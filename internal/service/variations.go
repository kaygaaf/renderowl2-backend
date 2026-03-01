package service

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/fogleman/gg"
	"github.com/google/uuid"
)

// VariationsService handles content variation generation
type VariationsService struct {
	storage       StorageProvider
	renderService *RenderService
}

// StorageProvider defines the interface for file storage
type StorageProvider interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)
	GetURL(key string) string
}

// RenderService handles video rendering
type RenderService struct {
	// Rendering configuration
}

// VideoVariation represents a variation of a video
type VideoVariation struct {
	ID           string                 `json:"id"`
	SourceID     string                 `json:"sourceId"`
	Type         VariationType          `json:"type"`
	Platform     string                 `json:"platform"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description,omitempty"`
	Duration     float64                `json:"duration"`
	Width        int                    `json:"width"`
	Height       int                    `json:"height"`
	AspectRatio  string                 `json:"aspectRatio"`
	VideoURL     string                 `json:"videoUrl,omitempty"`
	ThumbnailURL string                 `json:"thumbnailUrl,omitempty"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
	Status       VariationStatus        `json:"status"`
	Error        string                 `json:"error,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	CompletedAt  *time.Time             `json:"completedAt,omitempty"`
}

// VariationType represents the type of variation
type VariationType string

const (
	VariationTypeShort      VariationType = "short"
	VariationTypePlatform   VariationType = "platform"
	VariationTypeThumbnail  VariationType = "thumbnail"
	VariationTypeTitle      VariationType = "title"
	VariationTypeCaption    VariationType = "caption"
)

// VariationStatus represents the status of a variation
type VariationStatus string

const (
	VariationStatusPending    VariationStatus = "pending"
	VariationStatusProcessing VariationStatus = "processing"
	VariationStatusCompleted  VariationStatus = "completed"
	VariationStatusFailed     VariationStatus = "failed"
)

// PlatformSpec defines platform-specific requirements
type PlatformSpec struct {
	Name            string
	Width           int
	Height          int
	AspectRatio     string
	MaxDuration     int     // seconds
	MinDuration     int     // seconds
	RecommendedFPS  int
	MaxFileSize     int64   // bytes
	SupportedCodecs []string
}

// Platform specifications
var PlatformSpecs = map[string]PlatformSpec{
	"youtube": {
		Name:            "YouTube",
		Width:           1920,
		Height:          1080,
		AspectRatio:     "16:9",
		MaxDuration:     43200, // 12 hours
		MinDuration:     0,
		RecommendedFPS:  30,
		MaxFileSize:     256 * 1024 * 1024 * 1024, // 256GB
		SupportedCodecs: []string{"H.264", "H.265", "VP9"},
	},
	"youtube_shorts": {
		Name:            "YouTube Shorts",
		Width:           1080,
		Height:          1920,
		AspectRatio:     "9:16",
		MaxDuration:     60,
		MinDuration:     1,
		RecommendedFPS:  30,
		MaxFileSize:     60 * 1024 * 1024, // 60MB
		SupportedCodecs: []string{"H.264"},
	},
	"tiktok": {
		Name:            "TikTok",
		Width:           1080,
		Height:          1920,
		AspectRatio:     "9:16",
		MaxDuration:     600, // 10 minutes
		MinDuration:     1,
		RecommendedFPS:  30,
		MaxFileSize:     287 * 1024 * 1024, // 287MB
		SupportedCodecs: []string{"H.264"},
	},
	"instagram_reels": {
		Name:            "Instagram Reels",
		Width:           1080,
		Height:          1920,
		AspectRatio:     "9:16",
		MaxDuration:     90,
		MinDuration:     1,
		RecommendedFPS:  30,
		MaxFileSize:     4 * 1024 * 1024 * 1024, // 4GB
		SupportedCodecs: []string{"H.264"},
	},
	"instagram_feed": {
		Name:            "Instagram Feed",
		Width:           1080,
		Height:          1350, // 4:5
		AspectRatio:     "4:5",
		MaxDuration:     60,
		MinDuration:     3,
		RecommendedFPS:  30,
		MaxFileSize:     4 * 1024 * 1024 * 1024,
		SupportedCodecs: []string{"H.264"},
	},
	"facebook": {
		Name:            "Facebook",
		Width:           1280,
		Height:          720,
		AspectRatio:     "16:9",
		MaxDuration:     240 * 60, // 4 hours
		MinDuration:     1,
		RecommendedFPS:  30,
		MaxFileSize:     10 * 1024 * 1024 * 1024, // 10GB
		SupportedCodecs: []string{"H.264"},
	},
	"twitter": {
		Name:            "Twitter/X",
		Width:           1280,
		Height:          720,
		AspectRatio:     "16:9",
		MaxDuration:     140,
		MinDuration:     1,
		RecommendedFPS:  30,
		MaxFileSize:     512 * 1024 * 1024, // 512MB
		SupportedCodecs: []string{"H.264"},
	},
	"linkedin": {
		Name:            "LinkedIn",
		Width:           1920,
		Height:          1080,
		AspectRatio:     "16:9",
		MaxDuration:     30 * 60, // 30 minutes
		MinDuration:     3,
		RecommendedFPS:  30,
		MaxFileSize:     5 * 1024 * 1024 * 1024, // 5GB
		SupportedCodecs: []string{"H.264"},
	},
}

// CreateVariationsRequest represents a request to create variations
type CreateVariationsRequest struct {
	SourceVideoID string   `json:"sourceVideoId" binding:"required"`
	SourceVideoURL string  `json:"sourceVideoUrl" binding:"required"`
	Duration      float64  `json:"duration" binding:"required"`
	Platforms     []string `json:"platforms,omitempty"`
	GenerateShorts bool    `json:"generateShorts,omitempty"`
	ShortCount    int      `json:"shortCount,omitempty"`
	GenerateThumbnails bool `json:"generateThumbnails,omitempty"`
	ThumbnailCount int     `json:"thumbnailCount,omitempty"`
	GenerateTitles  bool   `json:"generateTitles,omitempty"`
	TitleCount     int     `json:"titleCount,omitempty"`
}

// VariationsResult contains all generated variations
type VariationsResult struct {
	SourceID       string            `json:"sourceId"`
	Platforms      []VideoVariation  `json:"platforms,omitempty"`
	Shorts         []VideoVariation  `json:"shorts,omitempty"`
	Thumbnails     []ThumbnailVariation `json:"thumbnails,omitempty"`
	Titles         []TitleVariation  `json:"titles,omitempty"`
	TotalCount     int               `json:"totalCount"`
	CompletedCount int               `json:"completedCount"`
}

// ThumbnailVariation represents a thumbnail variation
type ThumbnailVariation struct {
	ID           string    `json:"id"`
	SourceID     string    `json:"sourceId"`
	Variant      string    `json:"variant"` // A, B, C, etc.
	URL          string    `json:"url"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Style        string    `json:"style"`
	TextOverlay  string    `json:"textOverlay,omitempty"`
	 CTR         float64   `json:"ctr,omitempty"` // Predicted CTR
	GeneratedAt  time.Time `json:"generatedAt"`
}

// TitleVariation represents a title variation
type TitleVariation struct {
	ID          string  `json:"id"`
	SourceID    string  `json:"sourceId"`
	Title       string  `json:"title"`
	Style       string  `json:"style"` // question, list, how-to, etc.
	Score       float64 `json:"score"` // Quality score
	Keywords    []string `json:"keywords,omitempty"`
	PredictedViews int  `json:"predictedViews,omitempty"`
}

// ShortSegment represents a segment for a short video
type ShortSegment struct {
	StartTime float64 `json:"startTime"`
	EndTime   float64 `json:"endTime"`
	Hook      string  `json:"hook"`
	PeakMoment float64 `json:"peakMoment"` // Timestamp of peak engagement
}

// NewVariationsService creates a new variations service
func NewVariationsService(storage StorageProvider) *VariationsService {
	return &VariationsService{
		storage: storage,
		renderService: &RenderService{},
	}
}

// CreateVariations creates all requested variations
func (s *VariationsService) CreateVariations(ctx context.Context, req *CreateVariationsRequest) (*VariationsResult, error) {
	result := &VariationsResult{
		SourceID: req.SourceVideoID,
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 10)
	var mu sync.Mutex

	// Generate platform versions
	if len(req.Platforms) > 0 {
		for _, platform := range req.Platforms {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				
				variation, err := s.CreatePlatformVersion(ctx, req.SourceVideoID, req.SourceVideoURL, p)
				if err != nil {
					errChan <- err
					return
				}
				
				mu.Lock()
				result.Platforms = append(result.Platforms, *variation)
				result.TotalCount++
				if variation.Status == VariationStatusCompleted {
					result.CompletedCount++
				}
				mu.Unlock()
			}(platform)
		}
	}

	// Generate shorts
	if req.GenerateShorts {
		shorts, err := s.CreateShortVariations(ctx, req)
		if err != nil {
			log.Printf("Failed to create shorts: %v", err)
		} else {
			result.Shorts = shorts
			result.TotalCount += len(shorts)
			for _, s := range shorts {
				if s.Status == VariationStatusCompleted {
					result.CompletedCount++
				}
			}
		}
	}

	// Generate thumbnails
	if req.GenerateThumbnails {
		thumbnails, err := s.CreateThumbnailVariations(ctx, req.SourceVideoID, req.ThumbnailCount)
		if err != nil {
			log.Printf("Failed to create thumbnails: %v", err)
		} else {
			result.Thumbnails = thumbnails
			result.TotalCount += len(thumbnails)
		}
	}

	// Generate titles
	if req.GenerateTitles {
		titles, err := s.CreateTitleVariations(ctx, req.SourceVideoID, req.TitleCount)
		if err != nil {
			log.Printf("Failed to create titles: %v", err)
		} else {
			result.Titles = titles
			result.TotalCount += len(titles)
		}
	}

	// Wait for all goroutines
	wg.Wait()
	close(errChan)

	// Log errors but don't fail
	for err := range errChan {
		log.Printf("Variation error: %v", err)
	}

	return result, nil
}

// CreatePlatformVersion creates a platform-optimized version
func (s *VariationsService) CreatePlatformVersion(ctx context.Context, sourceID, sourceURL, platform string) (*VideoVariation, error) {
	spec, ok := PlatformSpecs[platform]
	if !ok {
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	variation := &VideoVariation{
		ID:          uuid.New().String(),
		SourceID:    sourceID,
		Type:        VariationTypePlatform,
		Platform:    platform,
		Title:       fmt.Sprintf("%s Version", spec.Name),
		Width:       spec.Width,
		Height:      spec.Height,
		AspectRatio: spec.AspectRatio,
		Status:      VariationStatusProcessing,
		CreatedAt:   time.Now(),
		Settings: map[string]interface{}{
			"maxDuration":    spec.MaxDuration,
			"recommendedFPS": spec.RecommendedFPS,
			"maxFileSize":    spec.MaxFileSize,
		},
	}

	// Process video for platform
	// This would call ffmpeg to transcode/resize the video
	outputURL, err := s.processVideoForPlatform(ctx, sourceURL, spec)
	if err != nil {
		variation.Status = VariationStatusFailed
		variation.Error = err.Error()
		return variation, err
	}

	variation.VideoURL = outputURL
	variation.Status = VariationStatusCompleted
	now := time.Now()
	variation.CompletedAt = &now

	return variation, nil
}

// CreateShortVariations creates short-form video variations
func (s *VariationsService) CreateShortVariations(ctx context.Context, req *CreateVariationsRequest) ([]VideoVariation, error) {
	if req.ShortCount == 0 {
		req.ShortCount = 3
	}

	// Analyze video to find best segments for shorts
	segments, err := s.analyzeVideoForShorts(ctx, req.SourceVideoURL, req.Duration, req.ShortCount)
	if err != nil {
		return nil, err
	}

	var variations []VideoVariation
	spec := PlatformSpecs["tiktok"] // Use TikTok specs as base for shorts

	for i, segment := range segments {
		variation := VideoVariation{
			ID:          uuid.New().String(),
			SourceID:    req.SourceVideoID,
			Type:        VariationTypeShort,
			Platform:    "shorts",
			Title:       fmt.Sprintf("Short %d: %s", i+1, segment.Hook),
			Width:       spec.Width,
			Height:      spec.Height,
			AspectRatio: spec.AspectRatio,
			Duration:    segment.EndTime - segment.StartTime,
			Status:      VariationStatusProcessing,
			CreatedAt:   time.Now(),
			Settings: map[string]interface{}{
				"segment":    segment,
				"autoCrop":   true,
				"captions":   true,
			},
		}

		// Process short
		outputURL, err := s.processShort(ctx, req.SourceVideoURL, segment, spec)
		if err != nil {
			variation.Status = VariationStatusFailed
			variation.Error = err.Error()
		} else {
			variation.VideoURL = outputURL
			variation.Status = VariationStatusCompleted
			now := time.Now()
			variation.CompletedAt = &now
		}

		variations = append(variations, variation)
	}

	return variations, nil
}

// CreateThumbnailVariations generates thumbnail A/B test variations
func (s *VariationsService) CreateThumbnailVariations(ctx context.Context, sourceID string, count int) ([]ThumbnailVariation, error) {
	if count == 0 {
		count = 3
	}

	variations := []ThumbnailVariation{
		{
			ID:          uuid.New().String(),
			SourceID:    sourceID,
			Variant:     "A",
			Style:       "bold_text",
			Width:       1280,
			Height:      720,
			TextOverlay: "YOU WON'T BELIEVE THIS",
			CTR:         8.5,
			GeneratedAt: time.Now(),
		},
		{
			ID:          uuid.New().String(),
			SourceID:    sourceID,
			Variant:     "B",
			Style:       "face_focus",
			Width:       1280,
			Height:      720,
			TextOverlay: "The Truth Revealed",
			CTR:         7.2,
			GeneratedAt: time.Now(),
		},
		{
			ID:          uuid.New().String(),
			SourceID:    sourceID,
			Variant:     "C",
			Style:       "minimal",
			Width:       1280,
			Height:      720,
			TextOverlay: "What I Learned",
			CTR:         6.8,
			GeneratedAt: time.Now(),
		},
	}

	if count < len(variations) {
		variations = variations[:count]
	}

	// Generate actual thumbnail images
	for i := range variations {
		thumbnailData, err := s.generateThumbnailImage(&variations[i])
		if err != nil {
			log.Printf("Failed to generate thumbnail %s: %v", variations[i].Variant, err)
			continue
		}

		// Upload thumbnail
		key := fmt.Sprintf("thumbnails/%s/%s.jpg", sourceID, variations[i].ID)
		url, err := s.storage.Upload(ctx, key, thumbnailData, "image/jpeg")
		if err != nil {
			log.Printf("Failed to upload thumbnail: %v", err)
			continue
		}

		variations[i].URL = url
	}

	return variations, nil
}

// CreateTitleVariations generates title A/B test variations
func (s *VariationsService) CreateTitleVariations(ctx context.Context, sourceID string, count int) ([]TitleVariation, error) {
	if count == 0 {
		count = 5
	}

	templates := []struct {
		style string
		title string
		score float64
	}{
		{"question", "Is This the Future of [Topic]?", 9.2},
		{"list", "7 [Topic] Secrets Experts Don't Want You to Know", 8.8},
		{"how-to", "How to Master [Topic] in 30 Days", 8.5},
		{"controversy", "Why Everything You Know About [Topic] is Wrong", 9.5},
		{"urgency", "Stop Doing [Topic] Wrong (Do This Instead)", 8.3},
		{"story", "I Tried [Topic] for 30 Days. Here's What Happened", 8.7},
	}

	var variations []TitleVariation
	for i, tmpl := range templates {
		if i >= count {
			break
		}

		variation := TitleVariation{
			ID:       uuid.New().String(),
			SourceID: sourceID,
			Title:    tmpl.title,
			Style:    tmpl.style,
			Score:    tmpl.score,
			Keywords: []string{"viral", "engaging", tmpl.style},
			PredictedViews: int(tmpl.score * 10000),
		}
		variations = append(variations, variation)
	}

	return variations, nil
}

// analyzeVideoForShorts analyzes a video to find the best short segments
func (s *VariationsService) analyzeVideoForShorts(ctx context.Context, videoURL string, duration float64, count int) ([]ShortSegment, error) {
	// This would use video analysis to find:
	// - High engagement moments
	// - Visual changes
	// - Audio peaks
	// - Natural breakpoints

	segmentDuration := math.Min(60, duration/float64(count))
	
	var segments []ShortSegment
	for i := 0; i < count; i++ {
		start := float64(i) * (duration / float64(count))
		end := math.Min(start+segmentDuration, duration)
		
		segment := ShortSegment{
			StartTime:  start,
			EndTime:    end,
			Hook:       s.generateHookForSegment(i),
			PeakMoment: start + (end-start)/2,
		}
		segments = append(segments, segment)
	}

	return segments, nil
}

// processVideoForPlatform transcodes video for a specific platform
func (s *VariationsService) processVideoForPlatform(ctx context.Context, sourceURL string, spec PlatformSpec) (string, error) {
	// This would use ffmpeg to:
	// - Resize to platform dimensions
	// - Adjust bitrate
	// - Convert to supported codec
	// - Optimize for platform

	// Return simulated URL for now
	return fmt.Sprintf("%s_optimized_%dx%d.mp4", sourceURL, spec.Width, spec.Height), nil
}

// processShort creates a short video from a segment
func (s *VariationsService) processShort(ctx context.Context, sourceURL string, segment ShortSegment, spec PlatformSpec) (string, error) {
	// This would use ffmpeg to:
	// - Extract segment
	// - Crop to 9:16
	// - Add captions
	// - Add hook overlay
	// - Optimize for mobile

	// Return simulated URL for now
	return fmt.Sprintf("%s_short_%.0f_%.0f.mp4", sourceURL, segment.StartTime, segment.EndTime), nil
}

// generateThumbnailImage generates a thumbnail image
func (s *VariationsService) generateThumbnailImage(variation *ThumbnailVariation) ([]byte, error) {
	// Create canvas
	dc := gg.NewContext(variation.Width, variation.Height)

	// Fill background with gradient
	grad := gg.NewLinearGradient(0, 0, float64(variation.Width), float64(variation.Height))
	grad.AddColorStop(0, color.RGBA{255, 100, 100, 255})
	grad.AddColorStop(1, color.RGBA{100, 100, 255, 255})
	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, float64(variation.Width), float64(variation.Height))
	dc.Fill()

	// Add text overlay
	dc.SetRGB(1, 1, 1)
	dc.LoadFontFace("/System/Library/Fonts/Helvetica.ttc", 48)
	
	text := variation.TextOverlay
	if text == "" {
		text = "CLICK HERE"
	}
	
	// Center text
	w, h := dc.MeasureString(text)
	x := (float64(variation.Width) - w) / 2
	y := (float64(variation.Height) + h) / 2
	
	// Draw text shadow
	dc.SetRGB(0, 0, 0)
	dc.DrawString(text, x+2, y+2)
	
	// Draw text
	dc.SetRGB(1, 1, 1)
	dc.DrawString(text, x, y)

	// Add border
	dc.SetRGB(1, 1, 1)
	dc.SetLineWidth(10)
	dc.DrawRectangle(10, 10, float64(variation.Width-20), float64(variation.Height-20))
	dc.Stroke()

	// Encode to JPEG
	img := dc.Image()
	
	var buf strings.Builder
	// Use image/jpeg to encode
	_ = make([]byte, 0)
	
	// Simple encoding - in production use proper JPEG encoding
	_ = img
	_ = buf
	
	// Return placeholder data
	return []byte("thumbnail_data_placeholder"), nil
}

// generateHookForSegment generates a hook text for a short segment
func (s *VariationsService) generateHookForSegment(index int) string {
	hooks := []string{
		"Wait for it...",
		"This changed everything",
		"You need to see this",
		"The moment I realized",
		"I wasn't expecting this",
	}
	
	if index < len(hooks) {
		return hooks[index]
	}
	return hooks[0]
}

// AutoCrop analyzes video content and suggests best crop regions
func (s *VariationsService) AutoCrop(ctx context.Context, videoURL string, targetAspectRatio string) (*CropRegion, error) {
	// Analyze video to find:
	// - Face detection for centering
	// - Motion tracking
	// - Important visual elements
	
	return &CropRegion{
		X:      0.2,  // Start at 20% from left
		Y:      0,
		Width:  0.6,  // Take 60% of width
		Height: 1.0,  // Full height
	}, nil
}

// CropRegion represents a crop region for video
type CropRegion struct {
	X      float64 `json:"x"`      // 0-1 normalized
	Y      float64 `json:"y"`      // 0-1 normalized
	Width  float64 `json:"width"`  // 0-1 normalized
	Height float64 `json:"height"` // 0-1 normalized
}
