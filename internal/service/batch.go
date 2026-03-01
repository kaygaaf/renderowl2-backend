package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"renderowl-api/internal/domain"
)

// Task types for batch processing
const (
	TypeBatchVideo   = "batch:video"
	TypeBatchProcess = "batch:process"
)

// BatchService manages batch video generation with queue processing
type BatchService struct {
	repo            domain.BatchRepository
	queue           *asynq.Client
	inspector       *asynq.Inspector
	timelineService *TimelineService
	clipService     *ClipService
	aiScriptService *AIScriptService
	aiSceneService  *AISceneService
	ttsService      *TTSService
	workerCount     int
}

// CreateBatchRequest represents a request to create a batch
type CreateBatchRequest struct {
	Name        string       `json:"name" binding:"required"`
	Description string       `json:"description,omitempty"`
	Videos      []VideoInput `json:"videos" binding:"required,min=1,max=30"`
	Config      domain.BatchConfig `json:"config" binding:"required"`
}

// VideoInput represents input for a single video
type VideoInput struct {
	Title       string            `json:"title" binding:"required"`
	Description string            `json:"description,omitempty"`
	Config      domain.VideoConfig `json:"config,omitempty"`
}

// BatchProgress represents the progress of a batch
type BatchProgress struct {
	BatchID      string  `json:"batchId"`
	Status       string  `json:"status"`
	Total        int     `json:"total"`
	Completed    int     `json:"completed"`
	Failed       int     `json:"failed"`
	InProgress   int     `json:"inProgress"`
	Progress     float64 `json:"progress"` // 0-100
	ETA          string  `json:"eta,omitempty"`
	CurrentVideo string  `json:"currentVideo,omitempty"`
}

// QueueStats represents queue statistics
type QueueStats struct {
	Pending   int `json:"pending"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Scheduled int `json:"scheduled"`
	Retry     int `json:"retry"`
}

// NewBatchService creates a new batch service
func NewBatchService(
	repo domain.BatchRepository,
	redisAddr string,
	redisPassword string,
	timelineService *TimelineService,
	clipService *ClipService,
	aiScriptService *AIScriptService,
	aiSceneService *AISceneService,
	ttsService *TTSService,
) (*BatchService, error) {
	queue := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	return &BatchService{
		repo:            repo,
		queue:           queue,
		inspector:       inspector,
		timelineService: timelineService,
		clipService:     clipService,
		aiScriptService: aiScriptService,
		aiSceneService:  aiSceneService,
		ttsService:      ttsService,
		workerCount:     3,
	}, nil
}

// CreateBatch creates a new batch job
func (s *BatchService) CreateBatch(ctx context.Context, userID string, req *CreateBatchRequest) (*domain.Batch, error) {
	batch := &domain.Batch{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Status:      domain.BatchStatusPending,
		TotalVideos: len(req.Videos),
		Config:      req.Config,
		Progress:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create batch videos
	for i, input := range req.Videos {
		video := domain.BatchVideo{
			ID:          uuid.New().String(),
			BatchID:     batch.ID,
			Title:       input.Title,
			Description: input.Description,
			Status:      domain.VideoStatusPending,
			Config:      input.Config,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Set order/index
		if video.Config.Topic == "" {
			video.Config.Topic = input.Title
		}

		batch.Videos = append(batch.Videos, video)

		// Update config defaults
		if i == 0 {
			if batch.Config.MaxConcurrent == 0 {
				batch.Config.MaxConcurrent = 3
			}
			if batch.Config.RetryAttempts == 0 {
				batch.Config.RetryAttempts = 2
			}
		}
	}

	// Save batch
	if err := s.repo.Create(batch); err != nil {
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}

	return batch, nil
}

// StartBatch starts processing a batch
func (s *BatchService) StartBatch(ctx context.Context, batchID string) error {
	batch, err := s.repo.Get(batchID)
	if err != nil {
		return err
	}

	if batch.Status != domain.BatchStatusPending {
		return fmt.Errorf("batch is not in pending status")
	}

	now := time.Now()
	batch.Status = domain.BatchStatusQueued
	batch.StartedAt = &now
	batch.UpdatedAt = now

	if err := s.repo.Update(batch); err != nil {
		return err
	}

	// Queue videos for processing
	for i := range batch.Videos {
		if err := s.queueVideo(&batch.Videos[i]); err != nil {
			log.Printf("Failed to queue video %s: %v", batch.Videos[i].ID, err)
			batch.Videos[i].Status = domain.VideoStatusFailed
			batch.Videos[i].Error = "Failed to queue"
		}
	}

	return nil
}

// queueVideo adds a video to the processing queue
func (s *BatchService) queueVideo(video *domain.BatchVideo) error {
	payload, err := json.Marshal(video)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeBatchVideo, payload)

	// Configure task options
	opts := []asynq.Option{
		asynq.Queue("batch"),
		asynq.MaxRetry(3),
		asynq.Timeout(30 * time.Minute),
		asynq.Retention(24 * time.Hour),
	}

	info, err := s.queue.Enqueue(task, opts...)
	if err != nil {
		return err
	}

	video.Status = domain.VideoStatusQueued
	log.Printf("Queued video %s with task ID %s", video.ID, info.ID)

	return nil
}

// GetBatch retrieves a batch by ID
func (s *BatchService) GetBatch(ctx context.Context, batchID string) (*domain.Batch, error) {
	return s.repo.Get(batchID)
}

// GetBatchProgress retrieves the current progress of a batch
func (s *BatchService) GetBatchProgress(ctx context.Context, batchID string) (*BatchProgress, error) {
	batch, err := s.repo.Get(batchID)
	if err != nil {
		return nil, err
	}

	// Calculate progress
	progress := &BatchProgress{
		BatchID:    batch.ID,
		Status:     string(batch.Status),
		Total:      batch.TotalVideos,
		Completed:  batch.Completed,
		Failed:     batch.Failed,
		InProgress: batch.InProgress,
		Progress:   batch.Progress,
	}

	// Find current video being processed
	for _, video := range batch.Videos {
		if video.Status == domain.VideoStatusProcessing {
			progress.CurrentVideo = video.Title
			break
		}
	}

	// Calculate ETA
	if batch.Status == domain.BatchStatusProcessing && batch.Completed > 0 {
		elapsed := time.Since(*batch.StartedAt)
		rate := float64(batch.Completed) / elapsed.Minutes()
		remaining := float64(batch.TotalVideos-batch.Completed) / rate
		progress.ETA = fmt.Sprintf("%.0f minutes", remaining)
	}

	return progress, nil
}

// GetBatchResults retrieves the results of a completed batch
func (s *BatchService) GetBatchResults(ctx context.Context, batchID string) (*domain.Batch, error) {
	batch, err := s.repo.Get(batchID)
	if err != nil {
		return nil, err
	}

	// Filter to only completed videos with results
	var completedVideos []domain.BatchVideo
	for _, video := range batch.Videos {
		if video.Status == domain.VideoStatusCompleted && video.Result != nil {
			completedVideos = append(completedVideos, video)
		}
	}

	batch.Videos = completedVideos
	return batch, nil
}

// ListBatches lists all batches for a user
func (s *BatchService) ListBatches(ctx context.Context, userID string, limit, offset int) ([]*domain.Batch, error) {
	if limit == 0 {
		limit = 20
	}
	return s.repo.List(userID, limit, offset)
}

// CancelBatch cancels a batch and all pending videos
func (s *BatchService) CancelBatch(ctx context.Context, batchID string) error {
	batch, err := s.repo.Get(batchID)
	if err != nil {
		return err
	}

	if batch.Status == domain.BatchStatusCompleted || batch.Status == domain.BatchStatusCancelled {
		return fmt.Errorf("batch cannot be cancelled")
	}

	batch.Status = domain.BatchStatusCancelled
	batch.UpdatedAt = time.Now()

	// Cancel all pending videos
	for i := range batch.Videos {
		if batch.Videos[i].Status == domain.VideoStatusPending || batch.Videos[i].Status == domain.VideoStatusQueued {
			batch.Videos[i].Status = domain.VideoStatusCancelled
		}
	}

	return s.repo.Update(batch)
}

// PauseBatch pauses batch processing
func (s *BatchService) PauseBatch(ctx context.Context, batchID string) error {
	batch, err := s.repo.Get(batchID)
	if err != nil {
		return err
	}

	if batch.Status != domain.BatchStatusProcessing {
		return fmt.Errorf("can only pause processing batches")
	}

	batch.Status = domain.BatchStatusPaused
	batch.UpdatedAt = time.Now()

	return s.repo.Update(batch)
}

// ResumeBatch resumes a paused batch
func (s *BatchService) ResumeBatch(ctx context.Context, batchID string) error {
	batch, err := s.repo.Get(batchID)
	if err != nil {
		return err
	}

	if batch.Status != domain.BatchStatusPaused {
		return fmt.Errorf("can only resume paused batches")
	}

	batch.Status = domain.BatchStatusProcessing
	batch.UpdatedAt = time.Now()

	return s.repo.Update(batch)
}

// RetryFailedVideos retries all failed videos in a batch
func (s *BatchService) RetryFailedVideos(ctx context.Context, batchID string) error {
	batch, err := s.repo.Get(batchID)
	if err != nil {
		return err
	}

	retryCount := 0
	for i := range batch.Videos {
		if batch.Videos[i].Status == domain.VideoStatusFailed {
			batch.Videos[i].Status = domain.VideoStatusPending
			batch.Videos[i].Error = ""
			batch.Videos[i].Progress = 0
			batch.Videos[i].UpdatedAt = time.Now()

			if err := s.queueVideo(&batch.Videos[i]); err != nil {
				log.Printf("Failed to requeue video %s: %v", batch.Videos[i].ID, err)
			} else {
				retryCount++
			}
		}
	}

	if retryCount > 0 {
		batch.Failed -= retryCount
		batch.UpdatedAt = time.Now()
		return s.repo.Update(batch)
	}

	return fmt.Errorf("no failed videos to retry")
}

// GetQueueStats retrieves queue statistics
func (s *BatchService) GetQueueStats(ctx context.Context) (*QueueStats, error) {
	stats, err := s.inspector.GetQueueInfo("batch")
	if err != nil {
		return nil, err
	}

	return &QueueStats{
		Pending:   stats.Pending,
		Active:    stats.Active,
		Completed: stats.Completed,
		Failed:    stats.Failed,
		Scheduled: stats.Scheduled,
		Retry:     stats.Retry,
	}, nil
}

// ProcessVideo processes a single video (called by worker)
func (s *BatchService) ProcessVideo(ctx context.Context, video *domain.BatchVideo) error {
	// Update status to processing
	video.Status = domain.VideoStatusProcessing
	now := time.Now()
	video.StartedAt = &now

	// Get batch for context
	batch, err := s.repo.Get(video.BatchID)
	if err != nil {
		return fmt.Errorf("failed to get batch: %w", err)
	}

	// Update batch progress
	batch.InProgress++
	batch.Status = domain.BatchStatusProcessing
	batch.UpdatedAt = time.Now()

	if err := s.repo.Update(batch); err != nil {
		return fmt.Errorf("failed to update batch: %w", err)
	}

	// Process the video
	result, err := s.generateVideo(ctx, video, batch)
	if err != nil {
		video.Status = domain.VideoStatusFailed
		video.Error = err.Error()
		video.Progress = 0

		// Update batch
		batch.InProgress--
		batch.Failed++
		batch.UpdatedAt = time.Now()
		s.repo.Update(batch)

		return err
	}

	// Success
	video.Status = domain.VideoStatusCompleted
	video.Result = result
	video.Progress = 100
	completedAt := time.Now()
	video.CompletedAt = &completedAt
	video.TimelineID = result.TimelineID

	// Update batch
	batch.InProgress--
	batch.Completed++
	batch.UpdatedAt = time.Now()

	// Calculate overall progress
	batch.Progress = float64(batch.Completed+batch.Failed) / float64(batch.TotalVideos) * 100

	// Check if batch is complete
	if batch.Completed+batch.Failed >= batch.TotalVideos {
		batch.CompletedAt = &completedAt
		if batch.Failed == 0 {
			batch.Status = domain.BatchStatusCompleted
		} else if batch.Completed > 0 {
			batch.Status = domain.BatchStatusCompleted // Partial success
		} else {
			batch.Status = domain.BatchStatusFailed
		}
	}

	return s.repo.Update(batch)
}

// generateVideo generates a single video
func (s *BatchService) generateVideo(ctx context.Context, video *domain.BatchVideo, batch *domain.Batch) (*domain.VideoResult, error) {
	startTime := time.Now()

	// Step 1: Generate script if not provided
	var script *Script
	if video.Config.Script != "" {
		script = &Script{
			Title:       video.Title,
			Description: video.Config.Script,
		}
	} else {
		scriptReq := &GenerateScriptRequest{
			Prompt:   video.Config.Topic,
			Style:    ScriptStyle(batch.Config.ScriptStyle),
			Tone:     video.Config.Tone,
			Duration: batch.Config.Duration,
		}

		var err error
		script, err = s.aiScriptService.GenerateScript(ctx, scriptReq)
		if err != nil {
			return nil, fmt.Errorf("script generation failed: %w", err)
		}
	}

	// Update progress
	video.Progress = 25
	s.repo.Update(batch)

	// Step 2: Generate scenes - convert script scenes to SceneInfo
	sceneInfos := make([]SceneInfo, 0, len(script.Scenes))
	for _, scene := range script.Scenes {
		sceneInfos = append(sceneInfos, SceneInfo{
			Number:      scene.Number,
			Title:       scene.Title,
			Description: scene.Description,
			Keywords:    scene.Keywords,
		})
	}

	sceneReq := &GenerateScenesRequest{
		ScriptID:       script.Title,
		Scenes:         sceneInfos,
		Style:          string(script.Style),
		ImageSource:    SourceUnsplash,
		GenerateImages: true,
	}

	scenes, err := s.aiSceneService.GenerateScenes(ctx, sceneReq)
	if err != nil {
		return nil, fmt.Errorf("scene generation failed: %w", err)
	}

	// Update progress
	video.Progress = 50
	s.repo.Update(batch)

	// Step 3: Generate voice if enabled
	if batch.Config.VoiceID != "" {
		ttsReq := &GenerateVoiceRequest{
			Text:           script.Description,
			VoiceID:        batch.Config.VoiceID,
			Provider:       ProviderElevenLabs,
			Speed:          1.0,
			ResponseFormat: "mp3",
		}

		_, err := s.ttsService.GenerateVoice(ctx, ttsReq)
		if err != nil {
			log.Printf("Voice generation failed for video %s: %v", video.ID, err)
			// Continue without voice - non-critical
		}
	}

	// Update progress
	video.Progress = 75
	s.repo.Update(batch)

	// Step 4: Create timeline
	timelineReq := &CreateTimelineRequest{
		Name:        video.Title,
		Description: video.Description,
		Duration:    float64(batch.Config.Duration),
		Width:       1920,
		Height:      1080,
		FPS:         30,
	}

	timeline, err := s.timelineService.Create(batch.UserID, timelineReq)
	if err != nil {
		return nil, fmt.Errorf("timeline creation failed: %w", err)
	}

	// Step 5: Add scenes as clips
	// Create a default track first or use the timeline ID as track reference
	currentTime := 0.0
	sceneDuration := float64(batch.Config.Duration) / float64(len(scenes.Scenes))
	if len(scenes.Scenes) == 0 {
		sceneDuration = 5.0 // Default 5 seconds per scene if no scenes
	}

	for i, scene := range scenes.Scenes {
		clipReq := &CreateClipRequest{
			TrackID:     timeline.ID, // Using timeline ID as track reference
			Name:        fmt.Sprintf("Scene %d: %s", i+1, scene.Title),
			Type:        "image",
			SourceURL:   scene.ImageURL,
			StartTime:   currentTime,
			EndTime:     currentTime + sceneDuration,
			TextContent: scene.Description,
		}

		_, err := s.clipService.Create(batch.UserID, timeline.ID, clipReq)
		if err != nil {
			log.Printf("Failed to add clip: %v", err)
		}
		currentTime += sceneDuration
	}

	renderTime := int(time.Since(startTime).Seconds())

	result := &domain.VideoResult{
		TimelineID: timeline.ID,
		Duration:   float64(batch.Config.Duration),
		Size:       0,
		Metadata:   map[string]string{"renderTime": fmt.Sprintf("%d", renderTime)},
	}

	return result, nil
}

// Close closes the batch service
func (s *BatchService) Close() error {
	return s.queue.Close()
}
