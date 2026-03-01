package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"renderowl-api/internal/middleware"
	"renderowl-api/internal/service"
)

// AIHandler handles AI-related HTTP requests
type AIHandler struct {
	scriptService *service.AIScriptService
	sceneService  *service.AISceneService
	ttsService    *service.TTSService
}

// NewAIHandler creates a new AI handler
func NewAIHandler(scriptService *service.AIScriptService, sceneService *service.AISceneService, ttsService *service.TTSService) *AIHandler {
	return &AIHandler{
		scriptService: scriptService,
		sceneService:  sceneService,
		ttsService:    ttsService,
	}
}

// GenerateScript generates a video script from a prompt
// POST /api/v1/ai/script
func (h *AIHandler) GenerateScript(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.GenerateScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	script, err := h.scriptService.GenerateScript(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "AI_GENERATION_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, script)
}

// EnhanceScript enhances an existing script
// POST /api/v1/ai/script/enhance
func (h *AIHandler) EnhanceScript(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Script          *service.Script `json:"script" binding:"required"`
		EnhancementType string          `json:"enhancement_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	script, err := h.scriptService.EnhanceScript(c.Request.Context(), req.Script, req.EnhancementType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "AI_ENHANCEMENT_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, script)
}

// GenerateScenes generates scenes from script information
// POST /api/v1/ai/scenes
func (h *AIHandler) GenerateScenes(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.GenerateScenesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	result, err := h.sceneService.GenerateScenes(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "AI_GENERATION_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GenerateVoice generates voice narration from text
// POST /api/v1/ai/voice
func (h *AIHandler) GenerateVoice(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.GenerateVoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	// Validate SSML if used
	if req.UseSSML {
		if err := h.ttsService.ValidateSSML(req.Text); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "SSML_VALIDATION_ERROR",
			})
			return
		}
	}

	result, err := h.ttsService.GenerateVoice(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "TTS_GENERATION_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListVoices returns available TTS voices
// GET /api/v1/ai/voices
func (h *AIHandler) ListVoices(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	voices, err := h.ttsService.ListVoices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "VOICE_LIST_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": voices,
		"meta": gin.H{
			"total": len(voices),
		},
	})
}

// GetScriptStyles returns available script styles
// GET /api/v1/ai/script-styles
func (h *AIHandler) GetScriptStyles(c *gin.Context) {
	styles := []map[string]interface{}{
		{
			"id":          service.StyleEducational,
			"name":        "Educational",
			"description": "Informative and instructional content",
			"icon":        "🎓",
		},
		{
			"id":          service.StyleEntertaining,
			"name":        "Entertaining",
			"description": "Engaging and fun content",
			"icon":        "🎬",
		},
		{
			"id":          service.StyleProfessional,
			"name":        "Professional",
			"description": "Business and corporate style",
			"icon":        "💼",
		},
		{
			"id":          service.StyleCasual,
			"name":        "Casual",
			"description": "Relaxed and conversational",
			"icon":        "☕",
		},
		{
			"id":          service.StyleDramatic,
			"name":        "Dramatic",
			"description": "Emotional and impactful",
			"icon":        "🎭",
		},
		{
			"id":          service.StyleHumorous,
			"name":        "Humorous",
			"description": "Funny and lighthearted",
			"icon":        "😄",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data": styles,
	})
}

// GetImageSources returns available image sources
// GET /api/v1/ai/image-sources
func (h *AIHandler) GetImageSources(c *gin.Context) {
	sources := []map[string]interface{}{
		{
			"id":           service.SourceUnsplash,
			"name":         "Unsplash",
			"description":  "High-quality free stock photos",
			"type":         "stock",
			"requires_key": true,
		},
		{
			"id":           service.SourcePexels,
			"name":         "Pexels",
			"description":  "Free stock photos and videos",
			"type":         "stock",
			"requires_key": true,
		},
		{
			"id":           service.SourceDALLE,
			"name":         "DALL-E 3",
			"description":  "AI-generated images via OpenAI",
			"type":         "ai",
			"requires_key": true,
		},
		{
			"id":           service.SourceStability,
			"name":         "Stability AI",
			"description":  "AI-generated images via Stability",
			"type":         "ai",
			"requires_key": true,
		},
		{
			"id":           service.SourceTogether,
			"name":         "Together AI",
			"description":  "AI-generated images via Together",
			"type":         "ai",
			"requires_key": true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data": sources,
	})
}
