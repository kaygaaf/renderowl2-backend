package social

import (
	"context"
	"renderowl-api/internal/domain/social"
)

// Platform defines the interface for all social media platforms
type Platform interface {
	// GetName returns the platform name
	GetName() social.SocialPlatform
	
	// GetAuthURL returns the OAuth URL for connecting an account
	GetAuthURL(state string) string
	
	// ExchangeCode exchanges OAuth code for tokens
	ExchangeCode(ctx context.Context, code string) (*social.SocialAccount, error)
	
	// RefreshToken refreshes the access token
	RefreshToken(ctx context.Context, account *social.SocialAccount) error
	
	// UploadVideo uploads a video to the platform
	UploadVideo(ctx context.Context, account *social.SocialAccount, req *social.UploadRequest) (*social.UploadResponse, error)
	
	// GetAnalytics retrieves analytics for a post
	GetAnalytics(ctx context.Context, account *social.SocialAccount, postID string) (*social.AnalyticsData, error)
	
	// DeletePost deletes a published post
	DeletePost(ctx context.Context, account *social.SocialAccount, postID string) error
	
	// GetTrends retrieves trending topics/sounds
	GetTrends(ctx context.Context, account *social.SocialAccount, region string) ([]*social.PlatformTrend, error)
}

// PlatformRegistry manages all available platforms
type PlatformRegistry struct {
	platforms map[social.SocialPlatform]Platform
}

// NewPlatformRegistry creates a new registry
func NewPlatformRegistry() *PlatformRegistry {
	return &PlatformRegistry{
		platforms: make(map[social.SocialPlatform]Platform),
	}
}

// Register adds a platform to the registry
func (r *PlatformRegistry) Register(platform Platform) {
	r.platforms[platform.GetName()] = platform
}

// Get retrieves a platform by name
func (r *PlatformRegistry) Get(name social.SocialPlatform) (Platform, bool) {
	p, ok := r.platforms[name]
	return p, ok
}

// GetAll returns all registered platforms
func (r *PlatformRegistry) GetAll() map[social.SocialPlatform]Platform {
	return r.platforms
}

// PlatformNames returns all platform names
func (r *PlatformRegistry) PlatformNames() []social.SocialPlatform {
	names := make([]social.SocialPlatform, 0, len(r.platforms))
	for name := range r.platforms {
		names = append(names, name)
	}
	return names
}
