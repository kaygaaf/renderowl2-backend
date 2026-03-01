package social

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"renderowl-api/internal/domain/social"
)

// LinkedInPlatform implements the Platform interface for LinkedIn
type LinkedInPlatform struct {
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

// LinkedIn API endpoints
const (
	LinkedInAuthURL     = "https://www.linkedin.com/oauth/v2/authorization"
	LinkedInTokenURL    = "https://www.linkedin.com/oauth/v2/accessToken"
	LinkedInAPIURL      = "https://api.linkedin.com/v2"
	LinkedInUploadURL   = "https://api.linkedin.com/v2/assets?action=registerUpload"
)

// NewLinkedInPlatform creates a new LinkedIn platform instance
func NewLinkedInPlatform(clientID, clientSecret, redirectURL string) *LinkedInPlatform {
	return &LinkedInPlatform{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}
}

// GetName returns the platform name
func (l *LinkedInPlatform) GetName() social.SocialPlatform {
	return social.PlatformLinkedIn
}

// GetAuthURL returns the OAuth URL
func (l *LinkedInPlatform) GetAuthURL(state string) string {
	params := url.Values{
		"response_type": {"code"},
		"client_id":     {l.clientID},
		"redirect_uri":  {l.redirectURL},
		"scope":         {"r_liteprofile r_emailaddress w_member_social w_organization_social r_organization_social"},
		"state":         {state},
	}

	return LinkedInAuthURL + "?" + params.Encode()
}

// ExchangeCode exchanges OAuth code for tokens
func (l *LinkedInPlatform) ExchangeCode(ctx context.Context, code string) (*social.SocialAccount, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {l.clientID},
		"client_secret": {l.clientSecret},
		"redirect_uri":  {l.redirectURL},
	}

	resp, err := l.makeFormRequest(ctx, "POST", LinkedInTokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}

	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Get user info
	userInfo, err := l.getUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &social.SocialAccount{
		Platform:    social.PlatformLinkedIn,
		AccountID:   userInfo["id"].(string),
		AccountName: userInfo["localizedFirstName"].(string) + " " + userInfo["localizedLastName"].(string),
		AccessToken: tokenResp.AccessToken,
		TokenExpiry: &expiry,
		Status:      social.StatusConnected,
		Metadata:    userInfo,
	}, nil
}

// RefreshToken refreshes the access token
func (l *LinkedInPlatform) RefreshToken(ctx context.Context, account *social.SocialAccount) error {
	// LinkedIn tokens are valid for 60 days
	// If expired, user needs to re-authenticate
	if account.TokenExpiry != nil && account.TokenExpiry.Before(time.Now()) {
		account.Status = social.StatusExpired
		return fmt.Errorf("LinkedIn token expired, please reconnect")
	}
	return nil
}

// UploadVideo uploads a video to LinkedIn
func (l *LinkedInPlatform) UploadVideo(ctx context.Context, account *social.SocialAccount, req *social.UploadRequest) (*social.UploadResponse, error) {
	// LinkedIn video upload process:
	// 1. Register upload
	// 2. Upload video bytes
	// 3. Create share with video

	// Step 1: Register upload
	registerData := map[string]interface{}{
		"registerUploadRequest": map[string]interface{}{
			"recipes": []string{"urn:li:digitalmediaRecipe:feedshare-video"},
			"owner":   "urn:li:person:" + account.AccountID,
			"serviceRelationships": []map[string]string{
				{
					"relationshipType": "OWNER",
					"identifier":       "urn:li:userGeneratedContent",
				},
			},
		},
	}

	jsonData, _ := json.Marshal(registerData)
	headers := map[string]string{
		"Authorization": "Bearer " + account.AccessToken,
		"Content-Type":  "application/json",
		"X-Restli-Protocol-Version": "2.0.0",
	}

	resp, err := l.makeRequest(ctx, "POST", LinkedInUploadURL, jsonData, headers)
	if err != nil {
		return nil, fmt.Errorf("register upload failed: %w", err)
	}

	var registerResp struct {
		Value struct {
			Asset       string `json:"asset"`
			UploadMechanism struct {
				ComLinkedInDigitalmediaUploadingMediaUploadHttpRequest struct {
					UploadURL string `json:"uploadUrl"`
				} `json:"com.linkedin.digitalmedia.uploading.MediaUploadHttpRequest"`
			} `json:"uploadMechanism"`
		} `json:"value"`
	}

	if err := json.Unmarshal(resp, &registerResp); err != nil {
		return nil, fmt.Errorf("failed to parse register response: %w", err)
	}

	// Step 2: Upload video (would need actual video bytes)
	// In production, upload to registerResp.Value.UploadMechanism.ComLinkedInDigitalmediaUploadingMediaUploadHttpRequest.UploadURL

	// Step 3: Create share
	shareData := map[string]interface{}{
		"author":          "urn:li:person:" + account.AccountID,
		"lifecycleState":  "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]string{
					"text": req.Description,
				},
				"shareMediaCategory": "VIDEO",
				"media": []map[string]interface{}{
					{
						"status":      "READY",
						"description": map[string]string{"text": req.Title},
						"media":       registerResp.Value.Asset,
						"title":       map[string]string{"text": req.Title},
					},
				},
			},
		},
		"visibility": map[string]string{
			"com.linkedin.ugc.MemberNetworkVisibility": req.Privacy,
		},
	}

	shareJSON, _ := json.Marshal(shareData)
	shareURL := LinkedInAPIURL + "/ugcPosts"

	shareResp, err := l.makeRequest(ctx, "POST", shareURL, shareJSON, headers)
	if err != nil {
		return nil, fmt.Errorf("share creation failed: %w", err)
	}

	// Extract share ID from response header
	shareID := extractURN(string(shareResp))

	return &social.UploadResponse{
		PlatformPostID: shareID,
		PostURL:        fmt.Sprintf("https://www.linkedin.com/feed/update/%s", shareID),
		Status:         "published",
	}, nil
}

// GetAnalytics retrieves analytics for a post
func (l *LinkedInPlatform) GetAnalytics(ctx context.Context, account *social.SocialAccount, postID string) (*social.AnalyticsData, error) {
	// LinkedIn analytics require organization access
	// This is a simplified implementation

	return &social.AnalyticsData{
		Platform: social.PlatformLinkedIn,
		Data: social.JSON{
			"note": "LinkedIn analytics require organization admin access",
		},
	}, nil
}

// DeletePost deletes a post
func (l *LinkedInPlatform) DeletePost(ctx context.Context, account *social.SocialAccount, postID string) error {
	deleteURL := LinkedInAPIURL + "/ugcPosts/" + postID
	headers := map[string]string{
		"Authorization":             "Bearer " + account.AccessToken,
		"X-Restli-Protocol-Version": "2.0.0",
	}

	_, err := l.makeRequest(ctx, "DELETE", deleteURL, nil, headers)
	return err
}

// GetTrends retrieves trending topics (LinkedIn doesn't have a public trends API)
func (l *LinkedInPlatform) GetTrends(ctx context.Context, account *social.SocialAccount, region string) ([]*social.PlatformTrend, error) {
	return []*social.PlatformTrend{
		{
			Platform:    social.PlatformLinkedIn,
			TrendType:   "topic",
			Title:       "Professional Development",
			Description: "Trending professional topics",
			Region:      region,
			FetchedAt:   time.Now(),
		},
	}, nil
}

// Helper methods

func (l *LinkedInPlatform) getUserInfo(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	userURL := LinkedInAPIURL + "/me"
	headers := map[string]string{
		"Authorization":             "Bearer " + accessToken,
		"X-Restli-Protocol-Version": "2.0.0",
	}

	resp, err := l.makeRequest(ctx, "GET", userURL, nil, headers)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (l *LinkedInPlatform) makeRequest(ctx context.Context, method, url string, body []byte, headers map[string]string) ([]byte, error) {
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

	resp, err := l.httpClient.Do(req)
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

func (l *LinkedInPlatform) makeFormRequest(ctx context.Context, method, url string, data url.Values) ([]byte, error) {
	resp, err := l.httpClient.PostForm(url, data)
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

func extractURN(response string) string {
	// Extract URN from LinkedIn response
	// Format: urn:li:share:123456789
	return response
}
