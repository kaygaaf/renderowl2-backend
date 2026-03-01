package social

import (
	"context"
	"fmt"
	"os"

	"renderowl-api/internal/domain/social"
)

// Service manages all social media operations
type Service struct {
	registry  *PlatformRegistry
	accounts  AccountRepository
	posts     PostRepository
	analytics AnalyticsRepository
}

// AccountRepository defines account storage operations
type AccountRepository interface {
	Create(ctx context.Context, account *social.SocialAccount) error
	GetByID(ctx context.Context, id string) (*social.SocialAccount, error)
	GetByUser(ctx context.Context, userID string) ([]*social.SocialAccount, error)
	GetByUserAndPlatform(ctx context.Context, userID string, platform social.SocialPlatform) (*social.SocialAccount, error)
	Update(ctx context.Context, account *social.SocialAccount) error
	Delete(ctx context.Context, id string) error
}

// PostRepository defines post storage operations
type PostRepository interface {
	Create(ctx context.Context, post *social.ScheduledPost) error
	GetByID(ctx context.Context, id string) (*social.ScheduledPost, error)
	GetByUser(ctx context.Context, userID string, limit, offset int) ([]*social.ScheduledPost, error)
	GetPending(ctx context.Context, before string) ([]*social.ScheduledPost, error)
	Update(ctx context.Context, post *social.ScheduledPost) error
	UpdateStatus(ctx context.Context, id string, status social.PostStatus, errorMsg string) error
	Delete(ctx context.Context, id string) error
}

// AnalyticsRepository defines analytics storage operations
type AnalyticsRepository interface {
	Create(ctx context.Context, data *social.AnalyticsData) error
	GetByPost(ctx context.Context, postID string) ([]*social.AnalyticsData, error)
	GetLatestByPost(ctx context.Context, postID string) (*social.AnalyticsData, error)
}

// NewService creates a new social media service
func NewService(
	registry *PlatformRegistry,
	accounts AccountRepository,
	posts PostRepository,
	analytics AnalyticsRepository,
) *Service {
	return &Service{
		registry:  registry,
		accounts:  accounts,
		posts:     posts,
		analytics: analytics,
	}
}

// InitializePlatforms sets up all platform instances with credentials from env
func (s *Service) InitializePlatforms() {
	// YouTube
	if clientID := os.Getenv("YOUTUBE_CLIENT_ID"); clientID != "" {
		yt := NewYouTubePlatform(
			clientID,
			os.Getenv("YOUTUBE_CLIENT_SECRET"),
			os.Getenv("YOUTUBE_REDIRECT_URL"),
		)
		s.registry.Register(yt)
	}

	// TikTok
	if clientKey := os.Getenv("TIKTOK_CLIENT_KEY"); clientKey != "" {
		tt := NewTikTokPlatform(
			clientKey,
			os.Getenv("TIKTOK_CLIENT_SECRET"),
			os.Getenv("TIKTOK_REDIRECT_URL"),
		)
		s.registry.Register(tt)
	}

	// Instagram
	if appID := os.Getenv("INSTAGRAM_APP_ID"); appID != "" {
		ig := NewInstagramPlatform(
			appID,
			os.Getenv("INSTAGRAM_APP_SECRET"),
			os.Getenv("INSTAGRAM_REDIRECT_URL"),
		)
		s.registry.Register(ig)
	}

	// Twitter/X
	if clientID := os.Getenv("TWITTER_CLIENT_ID"); clientID != "" {
		tw := NewTwitterPlatform(
			clientID,
			os.Getenv("TWITTER_CLIENT_SECRET"),
			os.Getenv("TWITTER_REDIRECT_URL"),
		)
		s.registry.Register(tw)
	}

	// LinkedIn
	if clientID := os.Getenv("LINKEDIN_CLIENT_ID"); clientID != "" {
		li := NewLinkedInPlatform(
			clientID,
			os.Getenv("LINKEDIN_CLIENT_SECRET"),
			os.Getenv("LINKEDIN_REDIRECT_URL"),
		)
		s.registry.Register(li)
	}

	// Facebook
	if appID := os.Getenv("FACEBOOK_APP_ID"); appID != "" {
		fb := NewFacebookPlatform(
			appID,
			os.Getenv("FACEBOOK_APP_SECRET"),
			os.Getenv("FACEBOOK_REDIRECT_URL"),
		)
		s.registry.Register(fb)
	}
}

// GetAuthURL returns the OAuth URL for a platform
func (s *Service) GetAuthURL(platform social.SocialPlatform, state string) (string, error) {
	p, ok := s.registry.Get(platform)
	if !ok {
		return "", fmt.Errorf("platform %s not configured", platform)
	}
	return p.GetAuthURL(state), nil
}

// ConnectAccount connects a new social media account
func (s *Service) ConnectAccount(ctx context.Context, platform social.SocialPlatform, code string, userID string) (*social.SocialAccount, error) {
	p, ok := s.registry.Get(platform)
	if !ok {
		return nil, fmt.Errorf("platform %s not configured", platform)
	}

	account, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	account.UserID = userID
	if err := s.accounts.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to save account: %w", err)
	}

	return account, nil
}

// GetAccounts returns all connected accounts for a user
func (s *Service) GetAccounts(ctx context.Context, userID string) ([]*social.SocialAccount, error) {
	return s.accounts.GetByUser(ctx, userID)
}

// GetAccount returns a specific account
func (s *Service) GetAccount(ctx context.Context, accountID string) (*social.SocialAccount, error) {
	return s.accounts.GetByID(ctx, accountID)
}

// DisconnectAccount removes a connected account
func (s *Service) DisconnectAccount(ctx context.Context, accountID string) error {
	return s.accounts.Delete(ctx, accountID)
}

// UploadVideo uploads a video to a platform
func (s *Service) UploadVideo(ctx context.Context, accountID string, req *social.UploadRequest) (*social.UploadResponse, error) {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	p, ok := s.registry.Get(account.Platform)
	if !ok {
		return nil, fmt.Errorf("platform %s not configured", account.Platform)
	}

	return p.UploadVideo(ctx, account, req)
}

// CrossPost uploads a video to multiple platforms
func (s *Service) CrossPost(ctx context.Context, accountIDs []string, req *social.UploadRequest) (map[string]*social.UploadResponse, error) {
	results := make(map[string]*social.UploadResponse)

	for _, accountID := range accountIDs {
		resp, err := s.UploadVideo(ctx, accountID, req)
		if err != nil {
			results[accountID] = &social.UploadResponse{
				Status: "failed",
			}
		} else {
			results[accountID] = resp
		}
	}

	return results, nil
}

// SchedulePost creates a scheduled post
func (s *Service) SchedulePost(ctx context.Context, post *social.ScheduledPost) error {
	// Validate all accounts exist and belong to user
	for _, platformPost := range post.Platforms {
		account, err := s.accounts.GetByID(ctx, platformPost.AccountID)
		if err != nil {
			return fmt.Errorf("account %s not found", platformPost.AccountID)
		}
		if account.UserID != post.UserID {
			return fmt.Errorf("account %s does not belong to user", platformPost.AccountID)
		}
	}

	post.Status = social.PostStatusScheduled
	return s.posts.Create(ctx, post)
}

// GetScheduledPosts returns scheduled posts for a user
func (s *Service) GetScheduledPosts(ctx context.Context, userID string, limit, offset int) ([]*social.ScheduledPost, error) {
	return s.posts.GetByUser(ctx, userID, limit, offset)
}

// CancelScheduledPost cancels a scheduled post
func (s *Service) CancelScheduledPost(ctx context.Context, postID string, userID string) error {
	post, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return err
	}
	if post.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	return s.posts.UpdateStatus(ctx, postID, social.PostStatusCancelled, "")
}

// GetAnalytics retrieves analytics for a post
func (s *Service) GetAnalytics(ctx context.Context, accountID string, postID string) (*social.AnalyticsData, error) {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	p, ok := s.registry.Get(account.Platform)
	if !ok {
		return nil, fmt.Errorf("platform %s not configured", account.Platform)
	}

	return p.GetAnalytics(ctx, account, postID)
}

// GetTrends retrieves trends for a platform
func (s *Service) GetTrends(ctx context.Context, accountID string, region string) ([]*social.PlatformTrend, error) {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	p, ok := s.registry.Get(account.Platform)
	if !ok {
		return nil, fmt.Errorf("platform %s not configured", account.Platform)
	}

	return p.GetTrends(ctx, account, region)
}

// GetPlatforms returns all configured platforms
func (s *Service) GetPlatforms() []social.SocialPlatform {
	return s.registry.PlatformNames()
}
