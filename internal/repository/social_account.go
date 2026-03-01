package repository

import (
	"context"

	"gorm.io/gorm"
	"renderowl-api/internal/domain/social"
)

// SocialAccountRepository implements social.AccountRepository
type SocialAccountRepository struct {
	db *gorm.DB
}

// NewSocialAccountRepository creates a new repository
func NewSocialAccountRepository(db *gorm.DB) *SocialAccountRepository {
	return &SocialAccountRepository{db: db}
}

// Create creates a new social account
func (r *SocialAccountRepository) Create(ctx context.Context, account *social.SocialAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

// GetByID gets account by ID
func (r *SocialAccountRepository) GetByID(ctx context.Context, id string) (*social.SocialAccount, error) {
	var account social.SocialAccount
	err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error
	return &account, err
}

// GetByUser gets accounts by user ID
func (r *SocialAccountRepository) GetByUser(ctx context.Context, userID string) ([]*social.SocialAccount, error) {
	var accounts []*social.SocialAccount
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error
	return accounts, err
}

// GetByUserAndPlatform gets account by user and platform
func (r *SocialAccountRepository) GetByUserAndPlatform(ctx context.Context, userID string, platform social.SocialPlatform) (*social.SocialAccount, error) {
	var account social.SocialAccount
	err := r.db.WithContext(ctx).Where("user_id = ? AND platform = ?", userID, platform).First(&account).Error
	return &account, err
}

// Update updates an account
func (r *SocialAccountRepository) Update(ctx context.Context, account *social.SocialAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// Delete deletes an account
func (r *SocialAccountRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&social.SocialAccount{}, "id = ?", id).Error
}
