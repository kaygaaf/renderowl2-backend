package domain

import "time"

// Batch represents a batch video generation job
type Batch struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"userId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      BatchStatus            `json:"status"`
	TotalVideos int                    `json:"totalVideos"`
	Completed   int                    `json:"completed"`
	Failed      int                    `json:"failed"`
	InProgress  int                    `json:"inProgress"`
	Videos      []BatchVideo           `json:"videos"`
	Config      BatchConfig            `json:"config"`
	Progress    float64                `json:"progress"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	StartedAt   *time.Time             `json:"startedAt,omitempty"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BatchStatus represents the status of a batch job
type BatchStatus string

const (
	BatchStatusPending    BatchStatus = "pending"
	BatchStatusQueued     BatchStatus = "queued"
	BatchStatusProcessing BatchStatus = "processing"
	BatchStatusCompleted  BatchStatus = "completed"
	BatchStatusFailed     BatchStatus = "failed"
	BatchStatusCancelled  BatchStatus = "cancelled"
	BatchStatusPaused     BatchStatus = "paused"
)

// BatchVideo represents a single video in a batch
type BatchVideo struct {
	ID          string            `json:"id"`
	BatchID     string            `json:"batchId"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      VideoStatus       `json:"status"`
	TimelineID  string            `json:"timelineId,omitempty"`
	Error       string            `json:"error,omitempty"`
	Progress    float64           `json:"progress"`
	Config      VideoConfig       `json:"config"`
	Result      *VideoResult      `json:"result,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	StartedAt   *time.Time        `json:"startedAt,omitempty"`
	CompletedAt *time.Time        `json:"completedAt,omitempty"`
}

// VideoStatus represents the status of a single video
type VideoStatus string

const (
	VideoStatusPending    VideoStatus = "pending"
	VideoStatusQueued     VideoStatus = "queued"
	VideoStatusProcessing VideoStatus = "processing"
	VideoStatusCompleted  VideoStatus = "completed"
	VideoStatusFailed     VideoStatus = "failed"
	VideoStatusSkipped    VideoStatus = "skipped"
	VideoStatusCancelled  VideoStatus = "cancelled"
)

// BatchConfig contains configuration for the batch
type BatchConfig struct {
	TemplateID             string                 `json:"templateId,omitempty"`
	ScriptStyle            string                 `json:"scriptStyle"`
	Duration               int                    `json:"duration"`
	VoiceID                string                 `json:"voiceId,omitempty"`
	BackgroundMusic        bool                   `json:"backgroundMusic"`
	AutoGenerateThumbnails bool                   `json:"autoGenerateThumbnails"`
	Platforms              []string               `json:"platforms,omitempty"`
	ParallelProcessing     bool                   `json:"parallelProcessing"`
	MaxConcurrent          int                    `json:"maxConcurrent"`
	RetryAttempts          int                    `json:"retryAttempts"`
	CustomSettings         map[string]interface{} `json:"customSettings,omitempty"`
}

// VideoConfig contains configuration for a single video
type VideoConfig struct {
	Topic          string   `json:"topic"`
	Script         string   `json:"script,omitempty"`
	Keywords       []string `json:"keywords,omitempty"`
	Tone           string   `json:"tone,omitempty"`
	TargetDuration int      `json:"targetDuration,omitempty"`
}

// VideoResult contains the result of video generation
type VideoResult struct {
	VideoURL   string            `json:"videoUrl"`
	Thumbnail  string            `json:"thumbnail,omitempty"`
	Duration   float64           `json:"duration"`
	Format     string            `json:"format"`
	Size       int64             `json:"size"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	TimelineID string            `json:"timelineId,omitempty"`
}

// BatchRepository defines the interface for batch data storage
type BatchRepository interface {
	Create(batch *Batch) error
	Get(id string) (*Batch, error)
	Update(batch *Batch) error
	List(userID string, limit, offset int) ([]*Batch, error)
	Delete(id string) error
}
