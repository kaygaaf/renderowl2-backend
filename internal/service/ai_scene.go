package service

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
)

// AISceneService handles AI-powered scene generation and image search
type AISceneService struct {
	openAIKey      string
	togetherKey    string
	stabilityKey   string
	unsplashKey    string
	pexelsKey      string
	openAIBaseURL  string
	httpClient     *http.Client
}

// ImageSource represents the source of an image
type ImageSource string

const (
	SourceUnsplash      ImageSource = "unsplash"
	SourcePexels        ImageSource = "pexels"
	SourceDALLE         ImageSource = "dalle"
	SourceStability     ImageSource = "stability"
	SourceTogether      ImageSource = "together"
)

// GenerateScenesRequest represents a scene generation request
type GenerateScenesRequest struct {
	ScriptID      string      `json:"script_id,omitempty"`
	ScriptTitle   string      `json:"script_title,omitempty"`
	Scenes        []SceneInfo `json:"scenes" binding:"required"`
	Style         string      `json:"style,omitempty"` // cinematic, animated, realistic, etc.
	ImageSource   ImageSource `json:"image_source,omitempty"`
	GenerateImages bool       `json:"generate_images,omitempty"`
}

// SceneInfo represents basic scene information for generation
type SceneInfo struct {
	Number      int      `json:"number"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords,omitempty"`
}

// GeneratedScene represents a fully generated scene
type GeneratedScene struct {
	Number        int         `json:"number"`
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	EnhancedDesc  string      `json:"enhanced_description,omitempty"`
	ImageURL      string      `json:"image_url,omitempty"`
	ThumbnailURL  string      `json:"thumbnail_url,omitempty"`
	ImageSource   ImageSource `json:"image_source"`
	ImagePrompt   string      `json:"image_prompt,omitempty"`
	AltText       string      `json:"alt_text,omitempty"`
	ColorPalette  []string    `json:"color_palette,omitempty"`
	Mood          string      `json:"mood,omitempty"`
}

// SceneGenerationResult represents the complete result
type SceneGenerationResult struct {
	ScriptID    string           `json:"script_id,omitempty"`
	Scenes      []GeneratedScene `json:"scenes"`
	TotalScenes int              `json:"total_scenes"`
}

// NewAISceneService creates a new AI scene service
func NewAISceneService() *AISceneService {
	return &AISceneService{
		openAIKey:     os.Getenv("OPENAI_API_KEY"),
		togetherKey:   os.Getenv("TOGETHER_API_KEY"),
		stabilityKey:  os.Getenv("STABILITY_API_KEY"),
		unsplashKey:   os.Getenv("UNSPLASH_ACCESS_KEY"),
		pexelsKey:     os.Getenv("PEXELS_API_KEY"),
		openAIBaseURL: getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// GenerateScenes generates enhanced scenes with images
func (s *AISceneService) GenerateScenes(ctx context.Context, req *GenerateScenesRequest) (*SceneGenerationResult, error) {
	if req.Style == "" {
		req.Style = "cinematic"
	}
	if req.ImageSource == "" {
		req.ImageSource = SourceUnsplash
	}

	result := &SceneGenerationResult{
		ScriptID: req.ScriptID,
		Scenes:   make([]GeneratedScene, 0, len(req.Scenes)),
	}

	for _, sceneInfo := range req.Scenes {
		scene := GeneratedScene{
			Number:      sceneInfo.Number,
			Title:       sceneInfo.Title,
			Description: sceneInfo.Description,
		}

		// Enhance scene description with AI
		enhancedDesc, imagePrompt, err := s.enhanceSceneDescription(ctx, sceneInfo, req.Style)
		if err == nil {
			scene.EnhancedDesc = enhancedDesc
			scene.ImagePrompt = imagePrompt
			scene.Mood = s.extractMood(enhancedDesc)
			scene.ColorPalette = s.extractColorPalette(enhancedDesc)
		}

		// Get image based on source
		if req.GenerateImages {
			switch req.ImageSource {
			case SourceDALLE:
				if s.openAIKey != "" {
					imageURL, err := s.generateImageWithDALLE(ctx, scene.ImagePrompt)
					if err == nil {
						scene.ImageURL = imageURL
						scene.ThumbnailURL = imageURL
						scene.ImageSource = SourceDALLE
					}
				}
			case SourceStability:
				if s.stabilityKey != "" {
					imageURL, err := s.generateImageWithStability(ctx, scene.ImagePrompt)
					if err == nil {
						scene.ImageURL = imageURL
						scene.ThumbnailURL = imageURL
						scene.ImageSource = SourceStability
					}
				}
			case SourceTogether:
				if s.togetherKey != "" {
					imageURL, err := s.generateImageWithTogether(ctx, scene.ImagePrompt)
					if err == nil {
						scene.ImageURL = imageURL
						scene.ThumbnailURL = imageURL
						scene.ImageSource = SourceTogether
					}
				}
			case SourceUnsplash:
				imageURL, thumbnailURL, altText, err := s.searchUnsplash(ctx, sceneInfo.Keywords)
				if err == nil {
					scene.ImageURL = imageURL
					scene.ThumbnailURL = thumbnailURL
					scene.AltText = altText
					scene.ImageSource = SourceUnsplash
				}
			case SourcePexels:
				imageURL, thumbnailURL, altText, err := s.searchPexels(ctx, sceneInfo.Keywords)
				if err == nil {
					scene.ImageURL = imageURL
					scene.ThumbnailURL = thumbnailURL
					scene.AltText = altText
					scene.ImageSource = SourcePexels
				}
			}
		}

		result.Scenes = append(result.Scenes, scene)
	}

	result.TotalScenes = len(result.Scenes)
	return result, nil
}

// enhanceSceneDescription uses AI to enhance scene descriptions
func (s *AISceneService) enhanceSceneDescription(ctx context.Context, scene SceneInfo, style string) (enhancedDesc, imagePrompt string, err error) {
	systemPrompt := fmt.Sprintf(`You are an expert cinematographer and visual designer specializing in %s style.

Enhance the scene description and create a detailed image generation prompt.

Respond ONLY with a JSON object:
{
  "enhanced_description": "Detailed visual description with mood, lighting, composition",
  "image_prompt": "Detailed prompt for AI image generation, 100-200 words, describing the visual scene",
  "mood": "emotional tone",
  "color_palette": ["#hex1", "#hex2", "#hex3"]
}`, style)

	userPrompt := fmt.Sprintf("Scene %d: %s\nOriginal description: %s\nKeywords: %v",
		scene.Number, scene.Title, scene.Description, scene.Keywords)

	// Try OpenAI first
	if s.openAIKey != "" {
		return s.enhanceWithOpenAI(ctx, systemPrompt, userPrompt)
	}
	if s.togetherKey != "" {
		return s.enhanceWithTogether(ctx, systemPrompt, userPrompt)
	}

	// Fallback to basic enhancement
	return scene.Description, fmt.Sprintf("%s style scene: %s", style, scene.Description), nil
}

// enhanceWithOpenAI enhances scene using OpenAI
func (s *AISceneService) enhanceWithOpenAI(ctx context.Context, systemPrompt, userPrompt string) (string, string, error) {
	requestBody := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.7,
		"max_tokens": 1000,
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonBody, _ := json.Marshal(requestBody)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", s.openAIBaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.openAIKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("API error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Choices) == 0 {
		return "", "", fmt.Errorf("no response")
	}

	var enhancement struct {
		EnhancedDesc string   `json:"enhanced_description"`
		ImagePrompt  string   `json:"image_prompt"`
		Mood         string   `json:"mood"`
		ColorPalette []string `json:"color_palette"`
	}

	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &enhancement); err != nil {
		return "", "", err
	}

	return enhancement.EnhancedDesc, enhancement.ImagePrompt, nil
}

// enhanceWithTogether enhances scene using Together AI
func (s *AISceneService) enhanceWithTogether(ctx context.Context, systemPrompt, userPrompt string) (string, string, error) {
	requestBody := map[string]interface{}{
		"model": "meta-llama/Llama-3.3-70B-Instruct-Turbo",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.7,
		"max_tokens": 1000,
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonBody, _ := json.Marshal(requestBody)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://api.together.xyz/v1/chat/completions", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.togetherKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("API error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Choices) == 0 {
		return "", "", fmt.Errorf("no response")
	}

	var enhancement struct {
		EnhancedDesc string   `json:"enhanced_description"`
		ImagePrompt  string   `json:"image_prompt"`
		Mood         string   `json:"mood"`
		ColorPalette []string `json:"color_palette"`
	}

	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &enhancement); err != nil {
		return "", "", err
	}

	return enhancement.EnhancedDesc, enhancement.ImagePrompt, nil
}

// generateImageWithDALLE generates an image using DALL-E
func (s *AISceneService) generateImageWithDALLE(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model": "dall-e-3",
		"prompt": prompt,
		"size": "1024x1024",
		"quality": "standard",
		"n": 1,
	}

	jsonBody, _ := json.Marshal(requestBody)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", s.openAIBaseURL+"/images/generations", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.openAIKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("DALL-E error: %s", string(body))
	}

	var result struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("no image generated")
	}

	return result.Data[0].URL, nil
}

// generateImageWithStability generates an image using Stability AI
func (s *AISceneService) generateImageWithStability(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"text_prompts": []map[string]interface{}{
			{"text": prompt, "weight": 1.0},
		},
		"cfg_scale": 7,
		"samples": 1,
		"steps": 30,
	}

	jsonBody, _ := json.Marshal(requestBody)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://api.stability.ai/v2beta/stable-image/generate/sd3", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.stabilityKey)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Stability error: %s", string(body))
	}

	var result struct {
		Image string `json:"image"` // Base64 encoded
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Note: In production, you'd upload this to S3 and return the URL
	return "data:image/png;base64," + result.Image, nil
}

// generateImageWithTogether generates an image using Together AI
func (s *AISceneService) generateImageWithTogether(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model": "black-forest-labs/FLUX.1-schnell",
		"prompt": prompt,
		"width": 1024,
		"height": 1024,
		"steps": 4,
		"n": 1,
	}

	jsonBody, _ := json.Marshal(requestBody)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://api.together.xyz/v1/images/generations", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.togetherKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Together error: %s", string(body))
	}

	var result struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("no image generated")
	}

	return result.Data[0].URL, nil
}

// searchUnsplash searches for images on Unsplash
func (s *AISceneService) searchUnsplash(ctx context.Context, keywords []string) (imageURL, thumbnailURL, altText string, err error) {
	if s.unsplashKey == "" {
		return "", "", "", fmt.Errorf("unsplash key not configured")
	}

	query := url.QueryEscape(joinKeywords(keywords))
	searchURL := fmt.Sprintf("https://api.unsplash.com/search/photos?query=%s&per_page=1&orientation=landscape", query)

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	httpReq.Header.Set("Authorization", "Client-ID "+s.unsplashKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("unsplash error: %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			URLs struct {
				Regular string `json:"regular"`
				Small   string `json:"small"`
			} `json:"urls"`
			AltDescription string `json:"alt_description"`
			Description    string `json:"description"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", "", err
	}

	if len(result.Results) == 0 {
		return "", "", "", fmt.Errorf("no images found")
	}

	photo := result.Results[0]
	alt := photo.AltDescription
	if alt == "" {
		alt = photo.Description
	}

	return photo.URLs.Regular, photo.URLs.Small, alt, nil
}

// searchPexels searches for images on Pexels
func (s *AISceneService) searchPexels(ctx context.Context, keywords []string) (imageURL, thumbnailURL, altText string, err error) {
	if s.pexelsKey == "" {
		return "", "", "", fmt.Errorf("pexels key not configured")
	}

	query := url.QueryEscape(joinKeywords(keywords))
	searchURL := fmt.Sprintf("https://api.pexels.com/v1/search?query=%s&per_page=1&orientation=landscape", query)

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	httpReq.Header.Set("Authorization", s.pexelsKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("pexels error: %d", resp.StatusCode)
	}

	var result struct {
		Photos []struct {
			Src struct {
				Large    string `json:"large"`
				Medium   string `json:"medium"`
				Portrait string `json:"portrait"`
			} `json:"src"`
			Alt string `json:"alt"`
		} `json:"photos"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", "", err
	}

	if len(result.Photos) == 0 {
		return "", "", "", fmt.Errorf("no images found")
	}

	photo := result.Photos[0]
	return photo.Src.Large, photo.Src.Medium, photo.Alt, nil
}

// extractMood extracts mood from enhanced description
func (s *AISceneService) extractMood(description string) string {
	// Simple extraction - in production, use AI or keyword matching
	moods := []string{"dramatic", "peaceful", "energetic", "mysterious", "romantic", "tense", "joyful"}
	for _, mood := range moods {
		if bytes.Contains([]byte(description), []byte(mood)) {
			return mood
		}
	}
	return "neutral"
}

// extractColorPalette extracts color palette from enhanced description
func (s *AISceneService) extractColorPalette(description string) []string {
	// Simple extraction - in production, use AI analysis
	return []string{"#1a1a2e", "#16213e", "#0f3460"} // Default cinematic palette
}

func joinKeywords(keywords []string) string {
	if len(keywords) == 0 {
		return "landscape"
	}
	result := ""
	for i, k := range keywords {
		if i > 0 {
			result += " "
		}
		result += k
	}
	return result
}
