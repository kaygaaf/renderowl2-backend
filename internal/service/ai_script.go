package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// AIScriptService handles AI-powered script generation
type AIScriptService struct {
	openAIKey     string
	togetherKey   string
	openAIBaseURL string
	httpClient    *http.Client
}

// ScriptStyle represents different script styles
type ScriptStyle string

const (
	StyleEducational ScriptStyle = "educational"
	StyleEntertaining ScriptStyle = "entertaining"
	StyleProfessional ScriptStyle = "professional"
	StyleCasual      ScriptStyle = "casual"
	StyleDramatic    ScriptStyle = "dramatic"
	StyleHumorous    ScriptStyle = "humorous"
)

// GenerateScriptRequest represents a script generation request
type GenerateScriptRequest struct {
	Prompt      string      `json:"prompt" binding:"required"`
	Style       ScriptStyle `json:"style,omitempty"`
	Duration    int         `json:"duration,omitempty"` // Target duration in seconds
	MaxScenes   int         `json:"max_scenes,omitempty"`
	Language    string      `json:"language,omitempty"` // ISO language code
	Tone        string      `json:"tone,omitempty"`
	TargetAudience string   `json:"target_audience,omitempty"`
}

// Script represents a generated video script
type Script struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	TotalDuration int    `json:"total_duration"` // in seconds
	Scenes      []Scene  `json:"scenes"`
	Style       ScriptStyle `json:"style"`
	Language    string   `json:"language"`
	Keywords    []string `json:"keywords,omitempty"`
}

// Scene represents a single scene in a script
type Scene struct {
	Number      int      `json:"number"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Narration   string   `json:"narration"`
	Duration    int      `json:"duration"` // in seconds
	VisualNotes string   `json:"visual_notes,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

// NewAIScriptService creates a new AI script service
func NewAIScriptService() *AIScriptService {
	return &AIScriptService{
		openAIKey:     os.Getenv("OPENAI_API_KEY"),
		togetherKey:   os.Getenv("TOGETHER_API_KEY"),
		openAIBaseURL: getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateScript generates a video script from a prompt
func (s *AIScriptService) GenerateScript(ctx context.Context, req *GenerateScriptRequest) (*Script, error) {
	if req.Style == "" {
		req.Style = StyleEducational
	}
	if req.Duration == 0 {
		req.Duration = 60
	}
	if req.MaxScenes == 0 {
		req.MaxScenes = 5
	}
	if req.Language == "" {
		req.Language = "en"
	}

	// Build the system prompt
	systemPrompt := s.buildSystemPrompt(req)
	
	// Build the user prompt
	userPrompt := fmt.Sprintf("Create a video script about: %s", req.Prompt)

	// Try OpenAI first, fall back to Together
	if s.openAIKey != "" {
		return s.generateWithOpenAI(ctx, systemPrompt, userPrompt, req)
	}
	if s.togetherKey != "" {
		return s.generateWithTogether(ctx, systemPrompt, userPrompt, req)
	}

	return nil, fmt.Errorf("no AI API key configured")
}

// buildSystemPrompt creates the system prompt for script generation
func (s *AIScriptService) buildSystemPrompt(req *GenerateScriptRequest) string {
	return fmt.Sprintf(`You are an expert video scriptwriter specializing in %s content.

Create a detailed video script with the following specifications:
- Target Duration: %d seconds
- Maximum Scenes: %d
- Style: %s
- Language: %s
- Tone: %s

Respond ONLY with a valid JSON object in this exact format:
{
  "title": "Compelling Video Title",
  "description": "Brief description of the video",
  "total_duration": %d,
  "scenes": [
    {
      "number": 1,
      "title": "Scene Title",
      "description": "What happens visually in this scene",
      "narration": "The actual narration/voiceover text",
      "duration": 15,
      "visual_notes": "Specific visual directions",
      "keywords": ["relevant", "search", "terms"]
    }
  ],
  "style": "%s",
  "language": "%s",
  "keywords": ["main", "keywords", "for", "video"]
}

Important:
- Scene durations must sum to approximately %d seconds
- Make narration engaging and natural-sounding
- Include specific visual directions
- Ensure the script flows logically`,
		req.Style,
		req.Duration,
		req.MaxScenes,
		req.Style,
		req.Language,
		req.Tone,
		req.Duration,
		req.Style,
		req.Language,
		req.Duration,
	)
}

// generateWithOpenAI generates a script using OpenAI API
func (s *AIScriptService) generateWithOpenAI(ctx context.Context, systemPrompt, userPrompt string, req *GenerateScriptRequest) (*Script, error) {
	requestBody := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.7,
		"max_tokens": 4000,
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.openAIBaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.openAIKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	var script Script
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &script); err != nil {
		return nil, fmt.Errorf("failed to parse script JSON: %w", err)
	}

	// Ensure language is set
	if script.Language == "" {
		script.Language = req.Language
	}
	if script.Style == "" {
		script.Style = req.Style
	}

	return &script, nil
}

// generateWithTogether generates a script using Together AI API
func (s *AIScriptService) generateWithTogether(ctx context.Context, systemPrompt, userPrompt string, req *GenerateScriptRequest) (*Script, error) {
	requestBody := map[string]interface{}{
		"model": "meta-llama/Llama-3.3-70B-Instruct-Turbo",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.7,
		"max_tokens": 4000,
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.together.xyz/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.togetherKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Together API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from Together AI")
	}

	var script Script
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &script); err != nil {
		return nil, fmt.Errorf("failed to parse script JSON: %w", err)
	}

	// Ensure language is set
	if script.Language == "" {
		script.Language = req.Language
	}
	if script.Style == "" {
		script.Style = req.Style
	}

	return &script, nil
}

// EnhanceScript takes an existing script and enhances it
func (s *AIScriptService) EnhanceScript(ctx context.Context, script *Script, enhancementType string) (*Script, error) {
	systemPrompt := fmt.Sprintf(`You are an expert script editor. Enhance the provided script by %s.

Respond with the complete enhanced script in the same JSON format.`, enhancementType)

	scriptJSON, _ := json.Marshal(script)
	userPrompt := fmt.Sprintf("Enhance this script:\n%s", string(scriptJSON))

	if s.openAIKey != "" {
		return s.generateWithOpenAI(ctx, systemPrompt, userPrompt, &GenerateScriptRequest{
			Style:    script.Style,
			Language: script.Language,
		})
	}
	if s.togetherKey != "" {
		return s.generateWithTogether(ctx, systemPrompt, userPrompt, &GenerateScriptRequest{
			Style:    script.Style,
			Language: script.Language,
		})
	}

	return nil, fmt.Errorf("no AI API key configured")
}

// EstimateDuration estimates the duration of narration text
func (s *AIScriptService) EstimateDuration(text string, wpm float64) int {
	if wpm == 0 {
		wpm = 150 // Average speaking rate
	}
	wordCount := len(bytes.Fields([]byte(text)))
	seconds := float64(wordCount) / wpm * 60
	return int(seconds + 0.5) // Round to nearest second
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
