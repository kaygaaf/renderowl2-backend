package social

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"renderowl-api/internal/domain/social"
)

// TwitterPlatform implements the Platform interface for Twitter/X
type TwitterPlatform struct {
	apiKey       string
	apiSecret    string
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

// Twitter API v2 endpoints
const (
	TwitterAuthURL      = "https://twitter.com/i/oauth2/authorize"
	TwitterTokenURL     = "https://api.twitter.com/2/oauth2/token"
	TwitterAPIURL       = "https://api.twitter.com/2"
	TwitterUploadURL    = "https://upload.twitter.com/1.1/media/upload.json"
)

// NewTwitterPlatform creates a new Twitter/X platform instance
func NewTwitterPlatform(clientID, clientSecret, redirectURL string) *TwitterPlatform {
	return &TwitterPlatform{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}
}

// GetName returns the platform name
func (t *TwitterPlatform) GetName() social.SocialPlatform {
	return social.PlatformTwitter
}

// GetAuthURL returns the OAuth 2.0 URL
func (t *TwitterPlatform) GetAuthURL(state string) string {
	codeChallenge := generateCodeChallenge()
	
	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {t.clientID},
		"redirect_uri":          {t.redirectURL},
		"scope":                 {"tweet.read tweet.write users.read media.write offline.access"},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
	}

	return TwitterAuthURL + "?" + params.Encode()
}

// ExchangeCode exchanges OAuth code for tokens
func (t *TwitterPlatform) ExchangeCode(ctx context.Context, code string) (*social.SocialAccount, error) {
	// Create Basic Auth header
	auth := base64.StdEncoding.EncodeToString([]byte(t.clientID + ":" + t.clientSecret))

	data := url.Values{
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"client_id":     {t.clientID},
		"redirect_uri":  {t.redirectURL},
		"code_verifier": {"challenge"}, // Should match the code_challenge
	}

	headers := map[string]string{
		"Authorization": "Basic " + auth,
		"Content-Type":  "application/x-www-form-urlencoded",
	}

	resp, err := t.makeRequest(ctx, "POST", TwitterTokenURL, []byte(data.Encode()), headers)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
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
		Platform:     social.PlatformTwitter,
		AccountID:    userInfo["id"].(string),
		AccountName:  userInfo["username"].(string),
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenExpiry:  &expiry,
		Status:       social.StatusConnected,
		Metadata:     userInfo,
	}, nil
}

// RefreshToken refreshes the access token
func (t *TwitterPlatform) RefreshToken(ctx context.Context, account *social.SocialAccount) error {
	auth := base64.StdEncoding.EncodeToString([]byte(t.clientID + ":" + t.clientSecret))

	data := url.Values{
		"refresh_token": {account.RefreshToken},
		"grant_type":    {"refresh_token"},
		"client_id":     {t.clientID},
	}

	headers := map[string]string{
		"Authorization": "Basic " + auth,
		"Content-Type":  "application/x-www-form-urlencoded",
	}

	resp, err := t.makeRequest(ctx, "POST", TwitterTokenURL, []byte(data.Encode()), headers)
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

// UploadVideo uploads a video to Twitter
func (t *TwitterPlatform) UploadVideo(ctx context.Context, account *social.SocialAccount, req *social.UploadRequest) (*social.UploadResponse, error) {
	// Refresh token if needed
	if account.TokenExpiry != nil && account.TokenExpiry.Before(time.Now()) {
		if err := t.RefreshToken(ctx, account); err != nil {
			return nil, err
		}
	}

	// Step 1: Upload media using chunked upload for videos
	mediaID, err := t.uploadVideoChunked(ctx, account, req.VideoPath)
	if err != nil {
		return nil, fmt.Errorf("video upload failed: %w", err)
	}

	// Step 2: Create tweet with media
	tweetURL := TwitterAPIURL + "/tweets"
	tweetData := map[string]interface{}{
		"text": req.Description,
		"media": map[string]interface{}{
			"media_ids": []string{mediaID},
		},
	}

	jsonData, _ := json.Marshal(tweetData)
	headers := map[string]string{
		"Authorization": "Bearer " + account.AccessToken,
		"Content-Type":  "application/json",
	}

	resp, err := t.makeRequest(ctx, "POST", tweetURL, jsonData, headers)
	if err != nil {
		return nil, fmt.Errorf("tweet creation failed: %w", err)
	}

	var tweetResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &tweetResp); err != nil {
		return nil, fmt.Errorf("failed to parse tweet response: %w", err)
	}

	return &social.UploadResponse{
		PlatformPostID: tweetResp.Data.ID,
		PostURL:        fmt.Sprintf("https://twitter.com/i/web/status/%s", tweetResp.Data.ID),
		Status:         "published",
	}, nil
}

// GetAnalytics retrieves analytics for a tweet
func (t *TwitterPlatform) GetAnalytics(ctx context.Context, account *social.SocialAccount, postID string) (*social.AnalyticsData, error) {
	// Twitter API v2 requires Elevated access for analytics
	// This is a simplified implementation

	return &social.AnalyticsData{
		Platform: social.PlatformTwitter,
		Data: social.JSON{
			"note": "Twitter analytics require Elevated API access",
		},
	}, nil
}

// DeletePost deletes a tweet
func (t *TwitterPlatform) DeletePost(ctx context.Context, account *social.SocialAccount, postID string) error {
	deleteURL := fmt.Sprintf("%s/tweets/%s", TwitterAPIURL, postID)
	headers := map[string]string{
		"Authorization": "Bearer " + account.AccessToken,
	}

	_, err := t.makeRequest(ctx, "DELETE", deleteURL, nil, headers)
	return err
}

// GetTrends retrieves trending topics
func (t *TwitterPlatform) GetTrends(ctx context.Context, account *social.SocialAccount, region string) ([]*social.PlatformTrend, error) {
	// Twitter API v2 trending is limited
	// Return mock data

	return []*social.PlatformTrend{
		{
			Platform:    social.PlatformTwitter,
			TrendType:   "hashtag",
			Title:       "#trending",
			Description: "What's happening now",
			Volume:      1000000,
			Region:      region,
			FetchedAt:   time.Now(),
		},
	}, nil
}

// Helper methods

func (t *TwitterPlatform) uploadVideoChunked(ctx context.Context, account *social.SocialAccount, videoPath string) (string, error) {
	file, err := os.Open(videoPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	fileSize := fileInfo.Size()

	// Step 1: INIT
	initParams := url.Values{
		"command":     {"INIT"},
		"media_type":  {"video/mp4"},
		"total_bytes": {fmt.Sprintf("%d", fileSize)},
	}

	headers := map[string]string{
		"Authorization": "Bearer " + account.AccessToken,
	}

	initResp, err := t.makeRequest(ctx, "POST", TwitterUploadURL+"?"+initParams.Encode(), nil, headers)
	if err != nil {
		return "", fmt.Errorf("upload init failed: %w", err)
	}

	var initResult struct {
		MediaID string `json:"media_id_string"`
	}
	if err := json.Unmarshal(initResp, &initResult); err != nil {
		return "", err
	}

	// Step 2: APPEND (simplified - would need chunking for large files)
	// For production, split into 5MB chunks
	buf := make([]byte, fileSize)
	file.Read(buf)

	appendParams := url.Values{
		"command":       {"APPEND"},
		"media_id":      {initResult.MediaID},
		"segment_index": {"0"},
	}

	// Use multipart form for video data
	// Simplified - production code would use proper multipart handling
	_, err = t.makeRequest(ctx, "POST", TwitterUploadURL+"?"+appendParams.Encode(), buf, headers)
	if err != nil {
		return "", fmt.Errorf("upload append failed: %w", err)
	}

	// Step 3: FINALIZE
	finalizeParams := url.Values{
		"command":  {"FINALIZE"},
		"media_id": {initResult.MediaID},
	}

	_, err = t.makeRequest(ctx, "POST", TwitterUploadURL+"?"+finalizeParams.Encode(), nil, headers)
	if err != nil {
		return "", fmt.Errorf("upload finalize failed: %w", err)
	}

	return initResult.MediaID, nil
}

func (t *TwitterPlatform) getUserInfo(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	userURL := TwitterAPIURL + "/users/me?user.fields=public_metrics,verified"
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	resp, err := t.makeRequest(ctx, "GET", userURL, nil, headers)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (t *TwitterPlatform) makeRequest(ctx context.Context, method, url string, body []byte, headers map[string]string) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func generateCodeChallenge() string {
	// Simplified - in production use PKCE properly
	return "challenge"
}
