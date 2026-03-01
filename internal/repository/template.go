package repository

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"renderowl-api/internal/domain"
)

// TemplateRepository defines template operations
type TemplateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// TemplateModel is the database model for templates
type TemplateModel struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `gorm:"not null"`
	Description string         `gorm:"not null"`
	Category    string         `gorm:"index;not null"`
	Thumbnail   string         `gorm:"not null"`
	Gradient    string         `gorm:"not null"`
	Icon        string         `gorm:"not null"`
	Duration    float64        `gorm:"default:60"`
	Width       int            `gorm:"default:1920"`
	Height      int            `gorm:"default:1080"`
	FPS         int            `gorm:"default:30"`
	Scenes      ScenesJSON     `gorm:"type:jsonb"`
	Popularity  int            `gorm:"default:0"`
	Version     int            `gorm:"default:1"`
	IsActive    bool           `gorm:"default:true;index"`
	Tags        TagsJSON       `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName specifies the table name for TemplateModel
func (TemplateModel) TableName() string {
	return "templates"
}

// ScenesJSON is a helper type for JSON serialization
type ScenesJSON []domain.TemplateScene

// Value implements driver.Valuer for database storage
func (s ScenesJSON) Value() (interface{}, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner for database retrieval
func (s *ScenesJSON) Scan(value interface{}) error {
	if value == nil {
		*s = ScenesJSON{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// TagsJSON is a helper type for JSON serialization of string slices
type TagsJSON []string

// Value implements driver.Valuer for database storage
func (t TagsJSON) Value() (interface{}, error) {
	return json.Marshal(t)
}

// Scan implements sql.Scanner for database retrieval
func (t *TagsJSON) Scan(value interface{}) error {
	if value == nil {
		*t = TagsJSON{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, t)
}

// Create creates a new template
func (r *TemplateRepository) Create(template *domain.Template) error {
	model := toTemplateModel(template)
	if err := r.db.Create(model).Error; err != nil {
		return err
	}
	*template = *fromTemplateModel(model)
	return nil
}

// GetByID retrieves a template by ID
func (r *TemplateRepository) GetByID(id string) (*domain.Template, error) {
	var model TemplateModel
	if err := r.db.First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}
	return fromTemplateModel(&model), nil
}

// List retrieves all templates with optional filtering
func (r *TemplateRepository) List(filter domain.TemplateFilter) ([]*domain.Template, error) {
	query := r.db.Where("is_active = ?", true)

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where(
			"name ILIKE ? OR description ILIKE ? OR tags @> ?",
			search, search, "[\""+filter.Search+"\"]",
		)
	}

	limit := filter.Limit
	if limit == 0 {
		limit = 50
	}

	var models []TemplateModel
	if err := query.Order("popularity DESC, created_at DESC").
		Limit(limit).
		Offset(filter.Offset).
		Find(&models).Error; err != nil {
		return nil, err
	}

	templates := make([]*domain.Template, len(models))
	for i, m := range models {
		templates[i] = fromTemplateModel(&m)
	}
	return templates, nil
}

// ListCategories retrieves all unique categories
func (r *TemplateRepository) ListCategories() ([]string, error) {
	var categories []string
	if err := r.db.Model(&TemplateModel{}).
		Where("is_active = ?", true).
		Distinct().
		Pluck("category", &categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// Update updates a template
func (r *TemplateRepository) Update(template *domain.Template) error {
	model := toTemplateModel(template)
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// Delete soft-deletes a template (marks as inactive)
func (r *TemplateRepository) Delete(id string) error {
	return r.db.Model(&TemplateModel{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// SeedDefaultTemplates seeds the database with default templates
func (r *TemplateRepository) SeedDefaultTemplates() error {
	defaults := domain.GetDefaultTemplates()

	for _, template := range defaults {
		// Check if template already exists
		var count int64
		r.db.Model(&TemplateModel{}).Where("id = ?", template.ID).Count(&count)

		if count == 0 {
			// Template doesn't exist, create it
			model := toTemplateModel(template)
			model.CreatedAt = time.Now()
			model.UpdatedAt = time.Now()
			if err := r.db.Create(model).Error; err != nil {
				return err
			}
		} else {
			// Template exists, update version if needed
			var existing TemplateModel
			r.db.First(&existing, "id = ?", template.ID)
			if existing.Version < template.Version {
				model := toTemplateModel(template)
				model.UpdatedAt = time.Now()
				r.db.Save(model)
			}
		}
	}

	return nil
}

// GetTemplateCount returns the total count of templates
func (r *TemplateRepository) GetTemplateCount() (int64, error) {
	var count int64
	err := r.db.Model(&TemplateModel{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

// Helper functions
func toTemplateModel(t *domain.Template) *TemplateModel {
	return &TemplateModel{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Category:    t.Category,
		Thumbnail:   t.Thumbnail,
		Gradient:    t.Gradient,
		Icon:        t.Icon,
		Duration:    t.Duration,
		Width:       t.Width,
		Height:      t.Height,
		FPS:         t.FPS,
		Scenes:      ScenesJSON(t.Scenes),
		Popularity:  t.Popularity,
		Version:     t.Version,
		IsActive:    t.IsActive,
		Tags:        TagsJSON(t.Tags),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func fromTemplateModel(m *TemplateModel) *domain.Template {
	return &domain.Template{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Category:    m.Category,
		Thumbnail:   m.Thumbnail,
		Gradient:    m.Gradient,
		Icon:        m.Icon,
		Duration:    m.Duration,
		Width:       m.Width,
		Height:      m.Height,
		FPS:         m.FPS,
		Scenes:      []domain.TemplateScene(m.Scenes),
		Popularity:  m.Popularity,
		Version:     m.Version,
		IsActive:    m.IsActive,
		Tags:        []string(m.Tags),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
