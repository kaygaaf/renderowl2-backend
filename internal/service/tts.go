package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// TTSService handles text-to-speech generation
type TTSService struct {
	elevenLabsKey string
	openAIKey     string
	httpClient    *http.Client
}

// TTSProvider represents the TTS provider
type TTSProvider string

const (
	ProviderElevenLabs TTSProvider = "elevenlabs"
	ProviderOpenAI     TTSProvider = "openai"
)

// Voice represents a TTS voice
type Voice struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Provider    TTSProvider       `json:"provider"`
	Gender      string            `json:"gender,omitempty"`
	Language    string            `json:"language,omitempty"`
	Accent      string            `json:"accent,omitempty"`
	Age         string            `json:"age,omitempty"`
	Description string            `json:"description,omitempty"`
	PreviewURL  string            `json:"preview_url,omitempty"`
	Category    string            `json:"category,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// GenerateVoiceRequest represents a voice generation request
type GenerateVoiceRequest struct {
	Text     string      `json:"text" binding:"required"`
	VoiceID  string      `json:"voice_id" binding:"required"`
	Provider TTSProvider `json:"provider,omitempty"`
	// ElevenLabs specific
	Stability       float64 `json:"stability,omitempty"`        // 0.0 - 1.0
	Clarity         float64 `json:"clarity,omitempty"`          // 0.0 - 1.0
	Style           float64 `json:"style,omitempty"`            // 0.0 - 1.0
	Speed           float64 `json:"speed,omitempty"`            // 0.5 - 2.0
	Model           string  `json:"model,omitempty"`            // eleven_multilingual_v2, etc.
	// OpenAI specific
	ResponseFormat  string  `json:"response_format,omitempty"`  // mp3, opus, aac, flac
	// SSML support
	UseSSML         bool    `json:"use_ssml,omitempty"`
}

// GenerateVoiceResponse represents the voice generation response
type GenerateVoiceResponse struct {
	AudioURL     string      `json:"audio_url,omitempty"`
	AudioBase64  string      `json:"audio_base64,omitempty"`
	Duration     float64     `json:"duration,omitempty"`
	Provider     TTSProvider `json:"provider"`
	VoiceID      string      `json:"voice_id"`
	Format       string      `json:"format"`
	Characters   int         `json:"characters"`
}

// SSMLBuilder helps build SSML content
type SSMLBuilder struct {
	strings.Builder
}

// NewSSMLBuilder creates a new SSML builder
func NewSSMLBuilder() *SSMLBuilder {
	b := &SSMLBuilder{}
	b.WriteString(`<speak>`)
	return b
}

// Close closes the SSML document
func (b *SSMLBuilder) Close() string {
	b.WriteString(`</speak>`)
	return b.String()
}

// AddPause adds a pause
func (b *SSMLBuilder) AddPause(duration string) {
	fmt.Fprintf(b, `<break time="%s"/>`, duration)
}

// AddEmphasis adds emphasis to text
func (b *SSMLBuilder) AddEmphasis(text, level string) {
	fmt.Fprintf(b, `<emphasis level="%s">%s</emphasis>`, level, text)
}

// AddProsody adds prosody (rate, pitch, volume)
func (b *SSMLBuilder) AddProsody(text, rate, pitch, volume string) {
	fmt.Fprintf(b, `<prosody`,)
	if rate != "" {
		fmt.Fprintf(b, ` rate="%s"`, rate)
	}
	if pitch != "" {
		fmt.Fprintf(b, ` pitch="%s"`, pitch)
	}
	if volume != "" {
		fmt.Fprintf(b, ` volume="%s"`, volume)
	}
	fmt.Fprintf(b, `>%s</prosody>`, text)
}

// AddParagraph adds a paragraph
func (b *SSMLBuilder) AddParagraph(text string) {
	fmt.Fprintf(b, `<p>%s</p>`, text)
}

// AddSentence adds a sentence
func (b *SSMLBuilder) AddSentence(text string) {
	fmt.Fprintf(b, `<s>%s</s>`, text)
}

// NewTTSService creates a new TTS service
func NewTTSService() *TTSService {
	return &TTSService{
		elevenLabsKey: os.Getenv("ELEVENLABS_API_KEY"),
		openAIKey:     os.Getenv("OPENAI_API_KEY"),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// ListVoices returns available voices from all configured providers
func (s *TTSService) ListVoices(ctx context.Context) ([]Voice, error) {
	var voices []Voice

	// Get ElevenLabs voices
	if s.elevenLabsKey != "" {
		elevens, err := s.listElevenLabsVoices(ctx)
		if err == nil {
			voices = append(voices, elevens...)
		}
	}

	// Get OpenAI voices
	if s.openAIKey != "" {
		openAIVoices := s.getOpenAIVoices()
		voices = append(voices, openAIVoices...)
	}

	return voices, nil
}

// listElevenLabsVoices fetches voices from ElevenLabs
func (s *TTSService) listElevenLabsVoices(ctx context.Context) ([]Voice, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", "https://api.elevenlabs.io/v1/voices", nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("xi-api-key", s.elevenLabsKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("elevenlabs error: %s", string(body))
	}

	var result struct {
		Voices []struct {
			VoiceID     string            `json:"voice_id"`
			Name        string            `json:"name"`
			Category    string            `json:"category"`
			Labels      map[string]string `json:"labels"`
			PreviewURL  string            `json:"preview_url"`
		} `json:"voices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	voices := make([]Voice, 0, len(result.Voices))
	for _, v := range result.Voices {
		voice := Voice{
			ID:         v.VoiceID,
			Name:       v.Name,
			Provider:   ProviderElevenLabs,
			Category:   v.Category,
			Labels:     v.Labels,
			PreviewURL: v.PreviewURL,
		}

		// Extract labels
		if v.Labels != nil {
			if gender, ok := v.Labels["gender"]; ok {
				voice.Gender = gender
			}
			if accent, ok := v.Labels["accent"]; ok {
				voice.Accent = accent
			}
			if age, ok := v.Labels["age"]; ok {
				voice.Age = age
			}
			if desc, ok := v.Labels["description"]; ok {
				voice.Description = desc
			}
		}

		voices = append(voices, voice)
	}

	return voices, nil
}

// getOpenAIVoices returns OpenAI's available voices
func (s *TTSService) getOpenAIVoices() []Voice {
	return []Voice{
		{ID: "alloy", Name: "Alloy", Provider: ProviderOpenAI, Gender: "neutral", Description: "Versatile, balanced voice"},
		{ID: "echo", Name: "Echo", Provider: ProviderOpenAI, Gender: "male", Description: "Warm, conversational male voice"},
		{ID: "fable", Name: "Fable", Provider: ProviderOpenAI, Gender: "male", Description: "British accent, storytelling quality"},
		{ID: "onyx", Name: "Onyx", Provider: ProviderOpenAI, Gender: "male", Description: "Deep, authoritative male voice"},
		{ID: "nova", Name: "Nova", Provider: ProviderOpenAI, Gender: "female", Description: "Energetic, expressive female voice"},
		{ID: "shimmer", Name: "Shimmer", Provider: ProviderOpenAI, Gender: "female", Description: "Clear, melodic female voice"},
	}
}

// GenerateVoice generates voice audio from text
func (s *TTSService) GenerateVoice(ctx context.Context, req *GenerateVoiceRequest) (*GenerateVoiceResponse, error) {
	// Set defaults
	if req.Provider == "" {
		if s.elevenLabsKey != "" {
			req.Provider = ProviderElevenLabs
		} else {
			req.Provider = ProviderOpenAI
		}
	}

	if req.Model == "" && req.Provider == ProviderElevenLabs {
		req.Model = "eleven_multilingual_v2"
	}

	if req.ResponseFormat == "" {
		req.ResponseFormat = "mp3"
	}

	switch req.Provider {
	case ProviderElevenLabs:
		return s.generateWithElevenLabs(ctx, req)
	case ProviderOpenAI:
		return s.generateWithOpenAI(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", req.Provider)
	}
}

// generateWithElevenLabs generates voice using ElevenLabs
func (s *TTSService) generateWithElevenLabs(ctx context.Context, req *GenerateVoiceRequest) (*GenerateVoiceResponse, error) {
	if s.elevenLabsKey == "" {
		return nil, fmt.Errorf("elevenlabs not configured")
	}

	requestBody := map[string]interface{}{
		"text": req.Text,
		"model_id": req.Model,
		"voice_settings": map[string]float64{
			"stability":       req.Stability,
			"similarity_boost": req.Clarity,
			"style":           req.Style,
			"speed":           req.Speed,
		},
	}

	// Set defaults
	if requestBody["voice_settings"].(map[string]float64)["stability"] == 0 {
		requestBody["voice_settings"].(map[string]float64)["stability"] = 0.5
	}
	if requestBody["voice_settings"].(map[string]float64)["similarity_boost"] == 0 {
		requestBody["voice_settings"].(map[string]float64)["similarity_boost"] = 0.75
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", req.VoiceID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("xi-api-key", s.elevenLabsKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("elevenlabs error: %s", string(body))
	}

	// Read audio data
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Estimate duration (rough estimate: ~150 words per minute)
	wordCount := len(bytes.Fields([]byte(req.Text)))
	duration := float64(wordCount) / 150.0 * 60.0

	return &GenerateVoiceResponse{
		AudioBase64: encodeBase64(audioData),
		Duration:    duration,
		Provider:    ProviderElevenLabs,
		VoiceID:     req.VoiceID,
		Format:      "mp3",
		Characters:  len(req.Text),
	}, nil
}

// generateWithOpenAI generates voice using OpenAI TTS
func (s *TTSService) generateWithOpenAI(ctx context.Context, req *GenerateVoiceRequest) (*GenerateVoiceResponse, error) {
	if s.openAIKey == "" {
		return nil, fmt.Errorf("openai not configured")
	}

	requestBody := map[string]interface{}{
		"model": "tts-1-hd",
		"input": req.Text,
		"voice": req.VoiceID,
	}

	if req.ResponseFormat != "" {
		requestBody["response_format"] = req.ResponseFormat
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/audio/speech", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.openAIKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai error: %s", string(body))
	}

	// Read audio data
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Estimate duration
	wordCount := len(bytes.Fields([]byte(req.Text)))
	duration := float64(wordCount) / 150.0 * 60.0

	return &GenerateVoiceResponse{
		AudioBase64: encodeBase64(audioData),
		Duration:    duration,
		Provider:    ProviderOpenAI,
		VoiceID:     req.VoiceID,
		Format:      req.ResponseFormat,
		Characters:  len(req.Text),
	}, nil
}

// CloneVoice creates a voice clone from audio samples (ElevenLabs only)
func (s *TTSService) CloneVoice(ctx context.Context, name string, description string, sampleFiles [][]byte) (*Voice, error) {
	if s.elevenLabsKey == "" {
		return nil, fmt.Errorf("elevenlabs not configured")
	}

	// Build multipart form
	body := &bytes.Buffer{}
	writer := bytes.NewBuffer(nil) // We'll use a different approach

	_ = body
	_ = writer

	// For now, return a placeholder - full implementation would use multipart form
	return &Voice{
		ID:          "cloned_" + name,
		Name:        name,
		Provider:    ProviderElevenLabs,
		Category:    "cloned",
		Description: description,
	}, fmt.Errorf("voice cloning not fully implemented")
}

// ValidateSSML validates SSML markup
func (s *TTSService) ValidateSSML(ssml string) error {
	if !strings.HasPrefix(ssml, "<speak>") {
		return fmt.Errorf("SSML must start with <speak>")
	}
	if !strings.HasSuffix(ssml, "</speak>") {
		return fmt.Errorf("SSML must end with </speak>")
	}
	return nil
}

func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
