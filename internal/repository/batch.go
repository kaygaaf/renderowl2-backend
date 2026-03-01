package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"renderowl-api/internal/domain"
)

// BatchRepository implements the batch repository interface
type BatchRepository struct {
	db *gorm.DB
}

// BatchModel is the database model for batches
type BatchModel struct {
	ID           string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       string `gorm:"index;not null"`
	Name         string `gorm:"not null"`
	Description  string
	Status       string  `gorm:"not null;default:'pending'"`
	TotalVideos  int     `gorm:"not null;default:0"`
	Completed    int     `gorm:"not null;default:0"`
	Failed       int     `gorm:"not null;default:0"`
	InProgress   int     `gorm:"not null;default:0"`
	Progress     float64 `gorm:"default:0"`
	Error        string
	ConfigJSON   string `gorm:"type:jsonb"`
	MetadataJSON string `gorm:"type:jsonb"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	StartedAt    *time.Time
	CompletedAt  *time.Time
	Videos       []BatchVideoModel `gorm:"foreignKey:BatchID;constraint:OnDelete:CASCADE;"`
}

// TableName specifies the table name
func (BatchModel) TableName() string {
	return "batches"
}

// BatchVideoModel is the database model for batch videos
type BatchVideoModel struct {
	ID          string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	BatchID     string  `gorm:"index;not null"`
	Title       string  `gorm:"not null"`
	Description string
	Status      string  `gorm:"not null;default:'pending'"`
	TimelineID  string
	Error       string
	Progress    float64 `gorm:"default:0"`
	ConfigJSON  string  `gorm:"type:jsonb"`
	ResultJSON  string  `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// TableName specifies the table name
func (BatchVideoModel) TableName() string {
	return "batch_videos"
}

// NewBatchRepository creates a new batch repository
func NewBatchRepository(db *gorm.DB) *BatchRepository {
	return &BatchRepository{db: db}
}

// Create creates a new batch
func (r *BatchRepository) Create(batch *domain.Batch) error {
	// Serialize config
	configJSON, err := json.Marshal(batch.Config)
	if err != nil {
		return err
	}

	// Serialize metadata
	metadataJSON, err := json.Marshal(batch.Metadata)
	if err != nil {
		return err
	}

	model := &BatchModel{
		ID:           batch.ID,
		UserID:       batch.UserID,
		Name:         batch.Name,
		Description:  batch.Description,
		Status:       string(batch.Status),
		TotalVideos:  batch.TotalVideos,
		Completed:    batch.Completed,
		Failed:       batch.Failed,
		InProgress:   batch.InProgress,
		Progress:     batch.Progress,
		Error:        batch.Error,
		ConfigJSON:   string(configJSON),
		MetadataJSON: string(metadataJSON),
		CreatedAt:    batch.CreatedAt,
		UpdatedAt:    batch.UpdatedAt,
		StartedAt:    batch.StartedAt,
		CompletedAt:  batch.CompletedAt,
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	// Create videos
	for _, video := range batch.Videos {
		if err := r.createVideo(&video); err != nil {
			return err
		}
	}

	return nil
}

// createVideo creates a batch video
func (r *BatchRepository) createVideo(video *domain.BatchVideo) error {
	configJSON, err := json.Marshal(video.Config)
	if err != nil {
		return err
	}

	var resultJSON []byte
	if video.Result != nil {
		resultJSON, err = json.Marshal(video.Result)
		if err != nil {
			return err
		}
	}

	model := &BatchVideoModel{
		ID:          video.ID,
		BatchID:     video.BatchID,
		Title:       video.Title,
		Description: video.Description,
		Status:      string(video.Status),
		TimelineID:  video.TimelineID,
		Error:       video.Error,
		Progress:    video.Progress,
		ConfigJSON:  string(configJSON),
		ResultJSON:  string(resultJSON),
		CreatedAt:   video.CreatedAt,
		UpdatedAt:   video.UpdatedAt,
		StartedAt:   video.StartedAt,
		CompletedAt: video.CompletedAt,
	}

	return r.db.Create(model).Error
}

// Get retrieves a batch by ID
func (r *BatchRepository) Get(id string) (*domain.Batch, error) {
	var model BatchModel
	if err := r.db.Preload("Videos").First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("batch not found")
		}
		return nil, err
	}

	return r.toDomain(&model), nil
}

// Update updates a batch
func (r *BatchRepository) Update(batch *domain.Batch) error {
	configJSON, err := json.Marshal(batch.Config)
	if err != nil {
		return err
	}

	metadataJSON, err := json.Marshal(batch.Metadata)
	if err != nil {
		return err
	}

	model := &BatchModel{
		ID:           batch.ID,
		UserID:       batch.UserID,
		Name:         batch.Name,
		Description:  batch.Description,
		Status:       string(batch.Status),
		TotalVideos:  batch.TotalVideos,
		Completed:    batch.Completed,
		Failed:       batch.Failed,
		InProgress:   batch.InProgress,
		Progress:     batch.Progress,
		Error:        batch.Error,
		ConfigJSON:   string(configJSON),
		MetadataJSON: string(metadataJSON),
		UpdatedAt:    batch.UpdatedAt,
		StartedAt:    batch.StartedAt,
		CompletedAt:  batch.CompletedAt,
	}

	if err := r.db.Save(model).Error; err != nil {
		return err
	}

	// Update videos
	for _, video := range batch.Videos {
		if err := r.updateVideo(&video); err != nil {
			return err
		}
	}

	return nil
}

// updateVideo updates a batch video
func (r *BatchRepository) updateVideo(video *domain.BatchVideo) error {
	configJSON, err := json.Marshal(video.Config)
	if err != nil {
		return err
	}

	var resultJSON []byte
	if video.Result != nil {
		resultJSON, err = json.Marshal(video.Result)
		if err != nil {
			return err
		}
	}

	model := &BatchVideoModel{
		ID:          video.ID,
		BatchID:     video.BatchID,
		Title:       video.Title,
		Description: video.Description,
		Status:      string(video.Status),
		TimelineID:  video.TimelineID,
		Error:       video.Error,
		Progress:    video.Progress,
		ConfigJSON:  string(configJSON),
		ResultJSON:  string(resultJSON),
		UpdatedAt:   video.UpdatedAt,
		StartedAt:   video.StartedAt,
		CompletedAt: video.CompletedAt,
	}

	return r.db.Save(model).Error
}

// List lists all batches for a user
func (r *BatchRepository) List(userID string, limit, offset int) ([]*domain.Batch, error) {
	var models []BatchModel
	if err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Videos").
		Find(&models).Error; err != nil {
		return nil, err
	}

	var batches []*domain.Batch
	for _, model := range models {
		batches = append(batches, r.toDomain(&model))
	}

	return batches, nil
}

// Delete deletes a batch
func (r *BatchRepository) Delete(id string) error {
	return r.db.Delete(&BatchModel{}, "id = ?", id).Error
}

// toDomain converts a model to domain object
func (r *BatchRepository) toDomain(model *BatchModel) *domain.Batch {
	// Deserialize config
	var config domain.BatchConfig
	json.Unmarshal([]byte(model.ConfigJSON), &config)

	// Deserialize metadata
	var metadata map[string]interface{}
	json.Unmarshal([]byte(model.MetadataJSON), &metadata)

	batch := &domain.Batch{
		ID:          model.ID,
		UserID:      model.UserID,
		Name:        model.Name,
		Description: model.Description,
		Status:      domain.BatchStatus(model.Status),
		TotalVideos: model.TotalVideos,
		Completed:   model.Completed,
		Failed:      model.Failed,
		InProgress:  model.InProgress,
		Progress:    model.Progress,
		Error:       model.Error,
		Config:      config,
		Metadata:    metadata,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		StartedAt:   model.StartedAt,
		CompletedAt: model.CompletedAt,
	}

	// Convert videos
	for _, videoModel := range model.Videos {
		batch.Videos = append(batch.Videos, *r.videoToDomain(&videoModel))
	}

	return batch
}

// videoToDomain converts a video model to domain object
func (r *BatchRepository) videoToDomain(model *BatchVideoModel) *domain.BatchVideo {
	// Deserialize config
	var config domain.VideoConfig
	json.Unmarshal([]byte(model.ConfigJSON), &config)

	// Deserialize result
	var result *domain.VideoResult
	if model.ResultJSON != "" {
		json.Unmarshal([]byte(model.ResultJSON), &result)
	}

	return &domain.BatchVideo{
		ID:          model.ID,
		BatchID:     model.BatchID,
		Title:       model.Title,
		Description: model.Description,
		Status:      domain.VideoStatus(model.Status),
		TimelineID:  model.TimelineID,
		Error:       model.Error,
		Progress:    model.Progress,
		Config:      config,
		Result:      result,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		StartedAt:   model.StartedAt,
		CompletedAt: model.CompletedAt,
	}
}

// Ensure BatchRepository implements the interface
type BatchRepositoryInterface interface {
	Create(batch *domain.Batch) error
	Get(id string) (*domain.Batch, error)
	Update(batch *domain.Batch) error
	List(userID string, limit, offset int) ([]*domain.Batch, error)
	Delete(id string) error
}

var _ BatchRepositoryInterface = (*BatchRepository)(nil)
