package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	Environment        string
	Port               string
	DatabaseURL        string
	RedisURL           string
	ClerkSecretKey     string
	FrontendURL        string
	RemotionURL        string
	// AI Service Keys
	OpenAIAPIKey       string
	TogetherAPIKey     string
	ElevenLabsAPIKey   string
	StabilityAPIKey    string
	UnsplashAccessKey  string
	PexelsAPIKey       string
	OpenAIBaseURL      string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Environment:       getEnv("ENVIRONMENT", "development"),
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/renderowl"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
		ClerkSecretKey:    getEnv("CLERK_SECRET_KEY", ""),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
		RemotionURL:       getEnv("REMOTION_URL", "http://localhost:3001"),
		// AI Service Keys
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		TogetherAPIKey:    getEnv("TOGETHER_API_KEY", ""),
		ElevenLabsAPIKey:  getEnv("ELEVENLABS_API_KEY", ""),
		StabilityAPIKey:   getEnv("STABILITY_API_KEY", ""),
		UnsplashAccessKey: getEnv("UNSPLASH_ACCESS_KEY", ""),
		PexelsAPIKey:      getEnv("PEXELS_API_KEY", ""),
		OpenAIBaseURL:     getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
