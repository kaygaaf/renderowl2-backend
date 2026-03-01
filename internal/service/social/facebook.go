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

// FacebookPlatform implements the Platform interface for Facebook
type FacebookPlatform struct {
	appID       string
	appSecret   string
	redirectURL string
	httpClient  *http.Client
}

// Facebook API endpoints
const (
	FacebookAuthURL   = "https://www.facebook.com/v18.0/dialog/oauth"
	FacebookTokenURL  = "https://graph.facebook.com/v18.0/oauth/access_token"
	FacebookGraphURL  = "https://graph.facebook.com/v18.0"
)

// NewFacebookPlatform creates a new Facebook platform instance
func NewFacebookPlatform(appID, appSecret, redirectURL string) *FacebookPlatform {
	return &FacebookPlatform{
		appID:       appID,
		appSecret:   appSecret,
		redirectURL: redirectURL,
		httpClient:  &http.Client{Timeout: 60 * time.Second},
	}
}

// GetName returns the platform name
func (f *FacebookPlatform) GetName() social.SocialPlatform {
	return social.PlatformFacebook
}

// GetAuthURL returns the OAuth URL
func (f *FacebookPlatform) GetAuthURL(state string) string {
	params := url.Values{
		"client_id":     {f.appID},
		"redirect_uri":  {f.redirectURL},
		"scope":         {"pages_manage_posts,pages_read_engagement,publish_video"},
		"response_type": {"code"},
		"state":         {state},
	}

	return FacebookAuthURL + "?" + params.Encode()
}

// ExchangeCode exchanges OAuth code for tokens
func (f *FacebookPlatform) ExchangeCode(ctx context.Context, code string) (*social.SocialAccount, error) {
	// Exchange code for access token
	data := url.Values{
		"client_id":     {f.appID},
		"client_secret": {f.appSecret},
		"code":          {code},
		"redirect_uri":  {f.redirectURL},
	}

	resp, err := f.makeRequest(ctx, "GET", FacebookTokenURL+"?"+data.Encode(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	var tokenResp struct {
		AccessToken string  `json:"access_token"`
		TokenType   string  `json:"token_type"`
		ExpiresIn   int     `json:"expires_in"`
	}

	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Get user's pages
	pages, err := f.getPages(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("no Facebook pages found")
	}

	// Use first page as default
	page := pages[0]

	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &social.SocialAccount{
		Platform:    social.PlatformFacebook,
		AccountID:   page["id"].(string),
		AccountName: page["name"].(string),
		AccessToken: page["access_token"].(string),
		TokenExpiry: &expiry,
		Status:      social.StatusConnected,
		Metadata: social.JSON{
			"page_category": page["category"],
			"pages":         pages,
		},
	}, nil
}

// RefreshToken refreshes the access token
func (f *FacebookPlatform) RefreshToken(ctx context.Context, account *social.SocialAccount) error {
	// Facebook long-lived tokens can be extended
	refreshURL := fmt.Sprintf("%s/oauth/access_token?grant_type=fb_exchange_token&client_id=%s&client_secret=%s&fb_exchange_token=%s",
		FacebookGraphURL, f.appID, f.appSecret, account.AccessToken)

	resp, err := f.makeRequest(ctx, "GET", refreshURL, nil, nil)
	if err != nil {
		account.Status = social.StatusExpired
		return fmt.Errorf("token refresh failed: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	account.AccessToken = tokenResp.AccessToken
	account.TokenExpiry = &expiry
	account.Status = social.StatusConnected

	return nil
}

// UploadVideo uploads a video to Facebook Page
func (f *FacebookPlatform) UploadVideo(ctx context.Context, account *social.SocialAccount, req *social.UploadRequest) (*social.UploadResponse, error) {
	// Facebook video upload
	uploadURL := fmt.Sprintf("%s/%s/videos", FacebookGraphURL, account.AccountID)

	params := url.Values{
		"access_token": {account.AccessToken},
		"file_url":     {req.VideoPath}, // Publicly accessible URL
		"title":        {req.Title},
		"description":  {req.Description},
	}

	if req.Privacy == "public" {
		params.Set("published", "true")
	} else {
		params.Set("published", "false")
	}

	resp, err := f.makeRequest(ctx, "POST", uploadURL+"?"+params.Encode(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("video upload failed: %w", err)
	}

	var uploadResp struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(resp, &uploadResp); err != nil {
		return nil, fmt.Errorf("failed to parse upload response: %w", err)
	}

	return &social.UploadResponse{
		PlatformPostID: uploadResp.ID,
		PostURL:        fmt.Sprintf("https://facebook.com/%s/videos/%s", account.AccountID, uploadResp.ID),
		Status:         "published",
	}, nil
}

// GetAnalytics retrieves analytics for a post
func (f *FacebookPlatform) GetAnalytics(ctx context.Context, account *social.SocialAccount, postID string) (*social.AnalyticsData, error) {
	insightsURL := fmt.Sprintf("%s/%s/insights?metric=post_impressions,post_engagements,post_video_views&access_token=%s",
		FacebookGraphURL, postID, account.AccessToken)

	resp, err := f.makeRequest(ctx, "GET", insightsURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("insights fetch failed: %w", err)
	}

	var insights struct {
		Data []struct {
			Name   string `json:"name"`
			Values []struct {
				Value int64 `json:"value"`
			} `json:"values"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &insights); err != nil {
		return nil, fmt.Errorf("failed to parse insights: %w", err)
	}

	analytics := &social.AnalyticsData{
		Platform: social.PlatformFacebook,
		Data:     make(social.JSON),
	}

	for _, metric := range insights.Data {
		if len(metric.Values) > 0 {
			switch metric.Name {
			case "post_impressions":
				analytics.Views = metric.Values[0].Value
			case "post_engagements":
				analytics.Engagement = float64(metric.Values[0].Value)
			case "post_video_views":
				analytics.Data["video_views"] = metric.Values[0].Value
			}
		}
	}

	return analytics, nil
}

// DeletePost deletes a post
func (f *FacebookPlatform) DeletePost(ctx context.Context, account *social.SocialAccount, postID string) error {
	deleteURL := fmt.Sprintf("%s/%s?access_token=%s", FacebookGraphURL, postID, account.AccessToken)
	_, err := f.makeRequest(ctx, "DELETE", deleteURL, nil, nil)
	return err
}

// GetTrends retrieves trending topics (Facebook doesn't have a public trends API)
func (f *FacebookPlatform) GetTrends(ctx context.Context, account *social.SocialAccount, region string) ([]*social.PlatformTrend, error) {
	return []*social.PlatformTrend{
		{
			Platform:    social.PlatformFacebook,
			TrendType:   "topic",
			Title:       "Trending Videos",
			Description: "Popular video content",
			Region:      region,
			FetchedAt:   time.Now(),
		},
	}, nil
}

// Helper methods

func (f *FacebookPlatform) getPages(ctx context.Context, userToken string) ([]map[string]interface{}, error) {
	pagesURL := fmt.Sprintf("%s/me/accounts?access_token=%s", FacebookGraphURL, userToken)

	resp, err := f.makeRequest(ctx, "GET", pagesURL, nil, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (f *FacebookPlatform) makeRequest(ctx context.Context, method, url string, body []byte, headers map[string]string) ([]byte, error) {
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

	resp, err := f.httpClient.Do(req)
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
