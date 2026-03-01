package service

import (
	"fmt"
	"time"

	"renderowl-api/internal/domain"
	"renderowl-api/internal/repository"
)

// TemplateService handles template business logic
type TemplateService struct {
	templateRepo *repository.TemplateRepository
	timelineRepo *repository.TimelineRepository
	trackRepo    *repository.TrackRepository
	clipRepo     *repository.ClipRepository
}

// NewTemplateService creates a new template service
func NewTemplateService(
	templateRepo *repository.TemplateRepository,
	timelineRepo *repository.TimelineRepository,
	trackRepo *repository.TrackRepository,
	clipRepo *repository.ClipRepository,
) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		timelineRepo: timelineRepo,
		trackRepo:    trackRepo,
		clipRepo:     clipRepo,
	}
}

// ListTemplates retrieves all templates with filtering
func (s *TemplateService) ListTemplates(filter domain.TemplateFilter) ([]*domain.Template, error) {
	return s.templateRepo.List(filter)
}

// GetTemplate retrieves a single template by ID
func (s *TemplateService) GetTemplate(id string) (*domain.Template, error) {
	return s.templateRepo.GetByID(id)
}

// GetCategories retrieves all template categories
func (s *TemplateService) GetCategories() ([]string, error) {
	return s.templateRepo.ListCategories()
}

// UseTemplate creates a new timeline from a template
func (s *TemplateService) UseTemplate(templateID, userID string, req *domain.UseTemplateRequest) (*domain.UseTemplateResponse, error) {
	// Get the template
	template, err := s.templateRepo.GetByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Create timeline name
	timelineName := req.Name
	if timelineName == "" {
		timelineName = template.Name + " - " + time.Now().Format("2006-01-02")
	}

	// Create the timeline
	timeline := &domain.Timeline{
		UserID:      userID,
		Name:        timelineName,
		Description: req.Description,
		Duration:    template.Duration,
		Width:       template.Width,
		Height:      template.Height,
		FPS:         template.FPS,
	}

	if err := s.timelineRepo.Create(timeline); err != nil {
		return nil, fmt.Errorf("failed to create timeline: %w", err)
	}

	// Create tracks and clips from template scenes
	if err := s.createTimelineFromTemplate(timeline.ID, template, req.CustomData); err != nil {
		// Clean up timeline if creation fails
		s.timelineRepo.Delete(timeline.ID)
		return nil, fmt.Errorf("failed to create timeline content: %w", err)
	}

	return &domain.UseTemplateResponse{
		TimelineID: timeline.ID,
		Message:    "Timeline created successfully from template",
		TemplateID: templateID,
		CreatedAt:  time.Now(),
	}, nil
}

// createTimelineFromTemplate creates tracks and clips based on template scenes
func (s *TemplateService) createTimelineFromTemplate(timelineID string, template *domain.Template, customData map[string]interface{}) error {
	// Create a video track
	videoTrack := &domain.Track{
		TimelineID: timelineID,
		Name:       "Video",
		Type:       "video",
		Order:      0,
		Muted:      false,
		Solo:       false,
	}
	if err := s.trackRepo.Create(videoTrack); err != nil {
		return err
	}

	// Create a text overlay track
	textTrack := &domain.Track{
		TimelineID: timelineID,
		Name:       "Text Overlays",
		Type:       "text",
		Order:      1,
		Muted:      false,
		Solo:       false,
	}
	if err := s.trackRepo.Create(textTrack); err != nil {
		return err
	}

	// Create clips from template scenes
	for _, scene := range template.Scenes {
		for _, templateClip := range scene.Clips {
			clip := s.templateClipToDomainClip(&templateClip, timelineID, videoTrack.ID, textTrack.ID, customData)
			if err := s.clipRepo.Create(clip); err != nil {
				return err
			}
		}
	}

	return nil
}

// templateClipToDomainClip converts a template clip to a domain clip
func (s *TemplateService) templateClipToDomainClip(
	tc *domain.TemplateClip,
	timelineID string,
	videoTrackID string,
	textTrackID string,
	customData map[string]interface{},
) *domain.Clip {
	clip := &domain.Clip{
		ID:          generateID(),
		TimelineID:  timelineID,
		TrackID:     videoTrackID,
		Name:        tc.Name,
		Type:        tc.Type,
		SourceURL:   tc.SourceURL,
		StartTime:   tc.StartTime,
		EndTime:     tc.EndTime,
		Duration:    tc.EndTime - tc.StartTime,
		TrimStart:   0,
		TrimEnd:     tc.EndTime - tc.StartTime,
		PositionX:   tc.PositionX,
		PositionY:   tc.PositionY,
		Scale:       tc.Scale,
		Rotation:    tc.Rotation,
		Opacity:     tc.Opacity,
	}

	// Determine track based on clip type
	if tc.Type == "text" {
		clip.TrackID = textTrackID
		clip.TextContent = s.replacePlaceholders(tc.TextContent, tc.Placeholder, customData)
		if tc.TextStyle != nil {
			clip.TextStyle = &domain.Style{
				FontSize:   tc.TextStyle.FontSize,
				FontFamily: tc.TextStyle.FontFamily,
				Color:      tc.TextStyle.Color,
				Background: tc.TextStyle.Background,
				Bold:       tc.TextStyle.Bold,
				Italic:     tc.TextStyle.Italic,
				Alignment:  tc.TextStyle.Alignment,
			}
		}
	}

	// Handle placeholder replacements for source URL
	if tc.SourceType == "placeholder" && tc.Placeholder != "" {
		if replacement, ok := customData[tc.Placeholder]; ok {
			if url, ok := replacement.(string); ok {
				clip.SourceURL = url
			}
		}
	}

	return clip
}

// replacePlaceholders replaces placeholder text with custom data
func (s *TemplateService) replacePlaceholders(text, placeholder string, customData map[string]interface{}) string {
	if placeholder == "" || customData == nil {
		return text
	}

	if replacement, ok := customData[placeholder]; ok {
		if str, ok := replacement.(string); ok {
			return str
		}
	}

	return text
}

// GetTemplateStats returns statistics about templates
func (s *TemplateService) GetTemplateStats() (*TemplateStats, error) {
	count, err := s.templateRepo.GetTemplateCount()
	if err != nil {
		return nil, err
	}

	categories, err := s.templateRepo.ListCategories()
	if err != nil {
		return nil, err
	}

	return &TemplateStats{
		TotalCount:     count,
		CategoryCount:  int64(len(categories)),
		Categories:     categories,
	}, nil
}

// TemplateStats represents template statistics
type TemplateStats struct {
	TotalCount     int64    `json:"totalCount"`
	CategoryCount  int64    `json:"categoryCount"`
	Categories     []string `json:"categories"`
}

// generateID generates a unique ID (placeholder - should use proper UUID generation)
func generateID() string {
	return fmt.Sprintf("clip_%d", time.Now().UnixNano())
}
