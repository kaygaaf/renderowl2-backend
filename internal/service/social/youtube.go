package social

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"renderowl-api/internal/domain/social"
)

// YouTubePlatform implements the Platform interface for YouTube
type YouTubePlatform struct {
	clientID     string
	clientSecret string
	redirectURL  string
	config       *oauth2.Config
}

// NewYouTubePlatform creates a new YouTube platform instance
func NewYouTubePlatform(clientID, clientSecret, redirectURL string) *YouTubePlatform {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			youtube.YoutubeUploadScope,
			youtube.YoutubeReadonlyScope,
			youtube.YoutubeScope,
		},
		Endpoint: google.Endpoint,
	}

	return &YouTubePlatform{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		config:       config,
	}
}

// GetName returns the platform name
func (y *YouTubePlatform) GetName() social.SocialPlatform {
	return social.PlatformYouTube
}

// GetAuthURL returns the OAuth URL
func (y *YouTubePlatform) GetAuthURL(state string) string {
	return y.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCode exchanges OAuth code for tokens
func (y *YouTubePlatform) ExchangeCode(ctx context.Context, code string) (*social.SocialAccount, error) {
	token, err := y.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info
	client := y.config.Client(ctx, token)
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Get channel info
	call := service.Channels.List([]string{"snippet", "contentDetails"})
	call = call.Mine(true)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get channel info: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("no YouTube channel found")
	}

	channel := response.Items[0]

	account := &social.SocialAccount{
		Platform:     social.PlatformYouTube,
		AccountID:    channel.Id,
		AccountName:  channel.Snippet.Title,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Status:       social.StatusConnected,
		Metadata: social.JSON{
			"channelThumbnail": channel.Snippet.Thumbnails.Default.Url,
		},
	}

	if !token.Expiry.IsZero() {
		account.TokenExpiry = &token.Expiry
	}

	return account, nil
}

// RefreshToken refreshes the access token
func (y *YouTubePlatform) RefreshToken(ctx context.Context, account *social.SocialAccount) error {
	token := &oauth2.Token{
		RefreshToken: account.RefreshToken,
	}

	tokenSource := y.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		account.Status = social.StatusExpired
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	account.AccessToken = newToken.AccessToken
	if !newToken.Expiry.IsZero() {
		account.TokenExpiry = &newToken.Expiry
	}
	account.Status = social.StatusConnected

	return nil
}

// UploadVideo uploads a video to YouTube
func (y *YouTubePlatform) UploadVideo(ctx context.Context, account *social.SocialAccount, req *social.UploadRequest) (*social.UploadResponse, error) {
	// Refresh token if needed
	if account.TokenExpiry != nil && account.TokenExpiry.Before(time.Now()) {
		if err := y.RefreshToken(ctx, account); err != nil {
			return nil, err
		}
	}

	token := &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
	}

	client := y.config.Client(ctx, token)
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Open video file
	file, err := os.Open(req.VideoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open video file: %w", err)
	}
	defer file.Close()

	// Create video snippet
	snippet := &youtube.VideoSnippet{
		Title:       req.Title,
		Description: req.Description,
	}

	if len(req.Tags) > 0 {
		snippet.Tags = req.Tags
	}

	// Set privacy status
	status := &youtube.VideoStatus{}
	switch req.Privacy {
	case "public":
		status.PrivacyStatus = "public"
	case "unlisted":
		status.PrivacyStatus = "unlisted"
	case "private":
		status.PrivacyStatus = "private"
	default:
		status.PrivacyStatus = "private"
	}

	// Create video resource
	video := &youtube.Video{
		Snippet: snippet,
		Status:  status,
	}

	// Upload
	call := service.Videos.Insert([]string{"snippet", "status"}, video)
	call = call.Media(file)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %w", err)
	}

	return &social.UploadResponse{
		PlatformPostID: response.Id,
		PostURL:        fmt.Sprintf("https://youtube.com/watch?v=%s", response.Id),
		Status:         "published",
	}, nil
}

// GetAnalytics retrieves analytics for a video
func (y *YouTubePlatform) GetAnalytics(ctx context.Context, account *social.SocialAccount, postID string) (*social.AnalyticsData, error) {
	// Refresh token if needed
	if account.TokenExpiry != nil && account.TokenExpiry.Before(time.Now()) {
		if err := y.RefreshToken(ctx, account); err != nil {
			return nil, err
		}
	}

	token := &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
	}

	client := y.config.Client(ctx, token)

	// Get video statistics
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	call := service.Videos.List([]string{"statistics"})
	call = call.Id(postID)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get video stats: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("video not found")
	}

	stats := response.Items[0].Statistics

	return &social.AnalyticsData{
		Platform:   social.PlatformYouTube,
		Views:      int64(stats.ViewCount),
		Likes:      int64(stats.LikeCount),
		Comments:   int64(stats.CommentCount),
		Data: social.JSON{
			"favoriteCount": int64(stats.FavoriteCount),
		},
	}, nil
}

// DeletePost deletes a video
func (y *YouTubePlatform) DeletePost(ctx context.Context, account *social.SocialAccount, postID string) error {
	// Refresh token if needed
	if account.TokenExpiry != nil && account.TokenExpiry.Before(time.Now()) {
		if err := y.RefreshToken(ctx, account); err != nil {
			return err
		}
	}

	token := &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
	}

	client := y.config.Client(ctx, token)
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create YouTube service: %w", err)
	}

	call := service.Videos.Delete(postID)
	return call.Do()
}

// GetTrends retrieves trending videos
func (y *YouTubePlatform) GetTrends(ctx context.Context, account *social.SocialAccount, region string) ([]*social.PlatformTrend, error) {
	token := &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
	}

	client := y.config.Client(ctx, token)
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	if region == "" {
		region = "US"
	}

	call := service.Videos.List([]string{"snippet", "statistics"})
	call = call.Chart("mostPopular").RegionCode(region).MaxResults(50)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get trending: %w", err)
	}

	trends := make([]*social.PlatformTrend, 0, len(response.Items))
	for _, video := range response.Items {
		trends = append(trends, &social.PlatformTrend{
			Platform:    social.PlatformYouTube,
			TrendType:   "video",
			Title:       video.Snippet.Title,
			Description: video.Snippet.Description,
			URL:         fmt.Sprintf("https://youtube.com/watch?v=%s", video.Id),
			Volume:      int64(video.Statistics.ViewCount),
			Region:      region,
			FetchedAt:   time.Now(),
		})
	}

	return trends, nil
}
