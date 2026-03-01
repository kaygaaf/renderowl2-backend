package social

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"renderowl-api/internal/domain/social"
)

// TikTokPlatform implements the Platform interface for TikTok
type TikTokPlatform struct {
	clientKey    string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

// TikTok API endpoints
const (
	TikTokAuthURL        = "https://www.tiktok.com/v2/auth/authorize/"
	TikTokTokenURL       = "https://open.tiktokapis.com/v2/oauth/token/"
	TikTokUploadURL      = "https://open.tiktokapis.com/v2/post/publish/video/init/"
	TikTokQueryUploadURL = "https://open.tiktokapis.com/v2/post/publish/video/status/"
	TikTokUserInfoURL    = "https://open.tiktokapis.com/v2/user/info/"
)

// NewTikTokPlatform creates a new TikTok platform instance
func NewTikTokPlatform(clientKey, clientSecret, redirectURL string) *TikTokPlatform {
	return &TikTokPlatform{
		clientKey:    clientKey,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// GetName returns the platform name
func (t *TikTokPlatform) GetName() social.SocialPlatform {
	return social.PlatformTikTok
}

// GetAuthURL returns the OAuth URL
func (t *TikTokPlatform) GetAuthURL(state string) string {
	params := map[string]string{
		"client_key":    t.clientKey,
		"redirect_uri":  t.redirectURL,
		"scope":         "video.publish,user.info.basic",
		"response_type": "code",
		"state":         state,
	}

	u, _ := url.Parse(TikTokAuthURL)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// ExchangeCode exchanges OAuth code for tokens
func (t *TikTokPlatform) ExchangeCode(ctx context.Context, code string) (*social.SocialAccount, error) {
	data := map[string]string{
		"client_key":    t.clientKey,
		"client_secret": t.clientSecret,
		"code":          code,
		"grant_type":    "authorization_code",
		"redirect_uri":  t.redirectURL,
	}

	resp, err := t.makeRequest(ctx, "POST", TikTokTokenURL, data, nil)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		OpenID       string `json:"open_id"`
		Scope        string `json:"scope"`
	}

	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Get user info
	userInfo, err := t.getUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &social.SocialAccount{
		Platform:     social.PlatformTikTok,
		AccountID:    tokenResp.OpenID,
		AccountName:  userInfo["display_name"].(string),
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenExpiry:  &expiry,
		Status:       social.StatusConnected,
		Metadata:     userInfo,
	}, nil
}

// RefreshToken refreshes the access token
func (t *TikTokPlatform) RefreshToken(ctx context.Context, account *social.SocialAccount) error {
	data := map[string]string{
		"client_key":    t.clientKey,
		"client_secret": t.clientSecret,
		"grant_type":    "refresh_token",
		"refresh_token": account.RefreshToken,
	}

	resp, err := t.makeRequest(ctx, "POST", TikTokTokenURL, data, nil)
	if err != nil {
		account.Status = social.StatusExpired
		return fmt.Errorf("token refresh failed: %w", err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	account.AccessToken = tokenResp.AccessToken
	account.RefreshToken = tokenResp.RefreshToken
	account.TokenExpiry = &expiry
	account.Status = social.StatusConnected

	return nil
}

// UploadVideo uploads a video to TikTok
func (t *TikTokPlatform) UploadVideo(ctx context.Context, account *social.SocialAccount, req *social.UploadRequest) (*social.UploadResponse, error) {
	// Refresh token if needed
	if account.TokenExpiry != nil && account.TokenExpiry.Before(time.Now()) {
		if err := t.RefreshToken(ctx, account); err != nil {
			return nil, err
		}
	}

	// Get file info (for future use with file size validation)
	_, err := os.Stat(req.VideoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat video file: %w", err)
	}
	
	// fileSize := fileInfo.Size() // Available for future size validation

	// Initialize upload
	initData := map[string]interface{}{
		"post_info": map[string]string{
			"title":       req.Title,
			"description": req.Description,
			"privacy_level": req.Privacy,
		},
		"source_info": map[string]interface{}{
			"source": "PULL_FROM_URL",
			"url":    req.VideoPath, // For direct upload, we'd need to host the file first
		},
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", account.AccessToken),
	}

	resp, err := t.makeRequest(ctx, "POST", TikTokUploadURL, initData, headers)
	if err != nil {
		return nil, fmt.Errorf("upload initialization failed: %w", err)
	}

	var uploadResp struct {
		Data struct {
			PublishID string `json:"publish_id"`
			UploadURL string `json:"upload_url"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &uploadResp); err != nil {
		return nil, fmt.Errorf("failed to parse upload response: %w", err)
	}

	// Note: TikTok requires the video to be uploaded to their URL
	// In production, we'd upload the file bytes to uploadResp.Data.UploadURL

	return &social.UploadResponse{
		PlatformPostID: uploadResp.Data.PublishID,
		PostURL:        fmt.Sprintf("https://tiktok.com/@%s/video/%s", account.AccountName, uploadResp.Data.PublishID),
		Status:         "processing",
	}, nil
}

// GetAnalytics retrieves analytics for a post
func (t *TikTokPlatform) GetAnalytics(ctx context.Context, account *social.SocialAccount, postID string) (*social.AnalyticsData, error) {
	// TikTok analytics API requires special permissions
	// This is a simplified implementation

	return &social.AnalyticsData{
		Platform: social.PlatformTikTok,
		Data: social.JSON{
			"note": "TikTok analytics require Creator Portal access",
		},
	}, nil
}

// DeletePost deletes a post (not directly supported by TikTok API)
func (t *TikTokPlatform) DeletePost(ctx context.Context, account *social.SocialAccount, postID string) error {
	return fmt.Errorf("TikTok does not support deleting posts via API")
}

// GetTrends retrieves trending sounds and hashtags
func (t *TikTokPlatform) GetTrends(ctx context.Context, account *social.SocialAccount, region string) ([]*social.PlatformTrend, error) {
	// TikTok Research API or unofficial endpoints would be needed
	// This returns mock data for demonstration

	trends := []*social.PlatformTrend{
		{
			Platform:    social.PlatformTikTok,
			TrendType:   "hashtag",
			Title:       "#viral",
			Description: "Trending viral content",
			Volume:      1000000000,
			Region:      region,
			FetchedAt:   time.Now(),
		},
		{
			Platform:    social.PlatformTikTok,
			TrendType:   "sound",
			Title:       "Popular Sound 2024",
			Description: "Trending audio track",
			Volume:      500000000,
			Region:      region,
			FetchedAt:   time.Now(),
		},
	}

	return trends, nil
}

// Helper methods

func (t *TikTokPlatform) getUserInfo(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}

	resp, err := t.makeRequest(ctx, "GET", TikTokUserInfoURL+"?fields=open_id,union_id,avatar_url,display_name", nil, headers)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			User map[string]interface{} `json:"user"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Data.User, nil
}

func (t *TikTokPlatform) makeRequest(ctx context.Context, method, url string, body interface{}, headers map[string]string) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
