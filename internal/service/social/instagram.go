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

// InstagramPlatform implements the Platform interface for Instagram
type InstagramPlatform struct {
	appID       string
	appSecret   string
	redirectURL string
	httpClient  *http.Client
}

// Instagram API endpoints
const (
	InstagramAuthURL     = "https://api.instagram.com/oauth/authorize"
	InstagramTokenURL    = "https://api.instagram.com/oauth/access_token"
	InstagramGraphURL    = "https://graph.instagram.com"
	InstagramGraphAPIURL = "https://graph.facebook.com/v18.0"
)

// NewInstagramPlatform creates a new Instagram platform instance
func NewInstagramPlatform(appID, appSecret, redirectURL string) *InstagramPlatform {
	return &InstagramPlatform{
		appID:       appID,
		appSecret:   appSecret,
		redirectURL: redirectURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// GetName returns the platform name
func (i *InstagramPlatform) GetName() social.SocialPlatform {
	return social.PlatformInstagram
}

// GetAuthURL returns the OAuth URL
func (i *InstagramPlatform) GetAuthURL(state string) string {
	params := url.Values{
		"client_id":     {i.appID},
		"redirect_uri":  {i.redirectURL},
		"scope":         {"instagram_basic,instagram_content_publish"},
		"response_type": {"code"},
		"state":         {state},
	}

	return InstagramAuthURL + "?" + params.Encode()
}

// ExchangeCode exchanges OAuth code for tokens
func (i *InstagramPlatform) ExchangeCode(ctx context.Context, code string) (*social.SocialAccount, error) {
	data := url.Values{
		"client_id":     {i.appID},
		"client_secret": {i.appSecret},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {i.redirectURL},
		"code":          {code},
	}

	resp, err := i.makeFormRequest(ctx, "POST", InstagramTokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		UserID      string `json:"user_id"`
	}

	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Get user info
	userInfo, err := i.getUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	return &social.SocialAccount{
		Platform:    social.PlatformInstagram,
		AccountID:   tokenResp.UserID,
		AccountName: userInfo["username"].(string),
		AccessToken: tokenResp.AccessToken,
		Status:      social.StatusConnected,
		Metadata:    userInfo,
	}, nil
}

// RefreshToken refreshes the access token
func (i *InstagramPlatform) RefreshToken(ctx context.Context, account *social.SocialAccount) error {
	// Instagram Basic Display API tokens can be refreshed
	refreshURL := fmt.Sprintf("%s/refresh_access_token?grant_type=ig_refresh_token&access_token=%s",
		InstagramGraphURL, account.AccessToken)

	resp, err := i.makeRequest(ctx, "GET", refreshURL, nil, nil)
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

// UploadVideo uploads a video to Instagram (as a Reel or post)
func (i *InstagramPlatform) UploadVideo(ctx context.Context, account *social.SocialAccount, req *social.UploadRequest) (*social.UploadResponse, error) {
	// Instagram requires videos to be hosted at a URL
	// For Reels API, we need to use the Facebook Graph API

	// Step 1: Create a media container
	createURL := fmt.Sprintf("%s/%s/media", InstagramGraphAPIURL, account.AccountID)
	params := url.Values{
		"media_type":   {"REELS"},
		"video_url":    {req.VideoPath}, // Must be a publicly accessible URL
		"caption":      {req.Description},
		"access_token": {account.AccessToken},
	}

	if len(req.Tags) > 0 {
		params.Set("share_to_feed", "true")
	}

	resp, err := i.makeRequest(ctx, "POST", createURL+"?"+params.Encode(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("media creation failed: %w", err)
	}

	var createResp struct {
		ID       string `json:"id"`
		MediaID  string `json:"media_id"`
		MediaURL string `json:"media_url"`
	}

	if err := json.Unmarshal(resp, &createResp); err != nil {
		return nil, fmt.Errorf("failed to parse creation response: %w", err)
	}

	// Step 2: Publish the container
	publishURL := fmt.Sprintf("%s/%s/media_publish", InstagramGraphAPIURL, account.AccountID)
	publishParams := url.Values{
		"creation_id":  {createResp.ID},
		"access_token": {account.AccessToken},
	}

	publishResp, err := i.makeRequest(ctx, "POST", publishURL+"?"+publishParams.Encode(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("media publish failed: %w", err)
	}

	var result struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(publishResp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse publish response: %w", err)
	}

	return &social.UploadResponse{
		PlatformPostID: result.ID,
		PostURL:        fmt.Sprintf("https://instagram.com/p/%s", result.ID),
		Status:         "published",
	}, nil
}

// GetAnalytics retrieves analytics for a post
func (i *InstagramPlatform) GetAnalytics(ctx context.Context, account *social.SocialAccount, postID string) (*social.AnalyticsData, error) {
	// Instagram Insights API
	insightsURL := fmt.Sprintf("%s/%s/insights?metric=engagement,impressions,reach,saved,video_views&access_token=%s",
		InstagramGraphAPIURL, postID, account.AccessToken)

	resp, err := i.makeRequest(ctx, "GET", insightsURL, nil, nil)
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
		Platform: social.PlatformInstagram,
		Data:     make(social.JSON),
	}

	for _, metric := range insights.Data {
		if len(metric.Values) > 0 {
			switch metric.Name {
			case "engagement":
				analytics.Engagement = float64(metric.Values[0].Value)
			case "impressions":
				analytics.Views = metric.Values[0].Value
			case "reach":
				analytics.Data["reach"] = metric.Values[0].Value
			case "saved":
				analytics.Data["saved"] = metric.Values[0].Value
			}
		}
	}

	return analytics, nil
}

// DeletePost deletes a post
func (i *InstagramPlatform) DeletePost(ctx context.Context, account *social.SocialAccount, postID string) error {
	deleteURL := fmt.Sprintf("%s/%s?access_token=%s", InstagramGraphAPIURL, postID, account.AccessToken)
	_, err := i.makeRequest(ctx, "DELETE", deleteURL, nil, nil)
	return err
}

// GetTrends retrieves trending hashtags
func (i *InstagramPlatform) GetTrends(ctx context.Context, account *social.SocialAccount, region string) ([]*social.PlatformTrend, error) {
	// Instagram doesn't provide a direct trending API
	// This would require scraping or third-party services

	return []*social.PlatformTrend{
		{
			Platform:  social.PlatformInstagram,
			TrendType: "hashtag",
			Title:     "#reels",
			Volume:    500000000,
			Region:    region,
			FetchedAt: time.Now(),
		},
	}, nil
}

// Helper methods

func (i *InstagramPlatform) getUserInfo(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	userURL := fmt.Sprintf("%s/me?fields=id,username,account_type,media_count&access_token=%s",
		InstagramGraphURL, accessToken)

	resp, err := i.makeRequest(ctx, "GET", userURL, nil, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (i *InstagramPlatform) makeRequest(ctx context.Context, method, url string, body []byte, headers map[string]string) ([]byte, error) {
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

	resp, err := i.httpClient.Do(req)
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

func (i *InstagramPlatform) makeFormRequest(ctx context.Context, method, url string, data url.Values) ([]byte, error) {
	resp, err := i.httpClient.PostForm(url, data)
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
