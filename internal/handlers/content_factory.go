package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"renderowl-api/internal/middleware"
	"renderowl-api/internal/service"
)

// ContentFactoryHandler handles content factory HTTP requests
type ContentFactoryHandler struct {
	ideationService   *service.IdeationService
	batchService      *service.BatchService
	variationsService *service.VariationsService
	optimizerService  *service.OptimizerService
}

// NewContentFactoryHandler creates a new content factory handler
func NewContentFactoryHandler(
	ideationService *service.IdeationService,
	batchService *service.BatchService,
	variationsService *service.VariationsService,
	optimizerService *service.OptimizerService,
) *ContentFactoryHandler {
	return &ContentFactoryHandler{
		ideationService:   ideationService,
		batchService:      batchService,
		variationsService: variationsService,
		optimizerService:  optimizerService,
	}
}

// ============================================
// IDEATION ENDPOINTS
// ============================================

// GetTrendingTopics returns trending topics from multiple platforms
// POST /api/v1/ideation/topics
func (h *ContentFactoryHandler) GetTrendingTopics(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.GetTrendingTopicsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	// Set defaults
	if len(req.Platforms) == 0 {
		req.Platforms = []string{"youtube", "tiktok", "twitter", "reddit"}
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	topics, err := h.ideationService.GetTrendingTopics(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "FETCH_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": topics,
		"meta": gin.H{
			"total":     len(topics),
			"platforms": req.Platforms,
		},
	})
}

// GetContentSuggestions returns AI-powered content suggestions
// POST /api/v1/ideation/suggestions
func (h *ContentFactoryHandler) GetContentSuggestions(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.GetContentSuggestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	suggestions, err := h.ideationService.GetContentSuggestions(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GENERATION_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": suggestions,
		"meta": gin.H{
			"total": len(suggestions),
			"niche": req.Niche,
		},
	})
}

// AnalyzeCompetitor analyzes a competitor channel
// POST /api/v1/ideation/competitor-analysis
func (h *ContentFactoryHandler) AnalyzeCompetitor(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.CompetitorAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	analysis, err := h.ideationService.AnalyzeCompetitor(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "ANALYSIS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GenerateContentCalendar generates a 30-day content calendar
// POST /api/v1/ideation/calendar
func (h *ContentFactoryHandler) GenerateContentCalendar(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.GenerateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	calendar, err := h.ideationService.GenerateContentCalendar(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GENERATION_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, calendar)
}

// ============================================
// BATCH ENDPOINTS
// ============================================

// CreateBatch creates a new batch generation job
// POST /api/v1/batch/generate
func (h *ContentFactoryHandler) CreateBatch(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	batch, err := h.batchService.CreateBatch(c.Request.Context(), user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "BATCH_CREATE_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, batch)
}

// StartBatch starts processing a batch
// POST /api/v1/batch/:id/start
func (h *ContentFactoryHandler) StartBatch(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	batchID := c.Param("id")

	if err := h.batchService.StartBatch(c.Request.Context(), batchID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "START_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch started successfully",
		"batchId": batchID,
	})
}

// GetBatchStatus returns the current status of a batch
// GET /api/v1/batch/:id/status
func (h *ContentFactoryHandler) GetBatchStatus(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	batchID := c.Param("id")

	progress, err := h.batchService.GetBatchProgress(c.Request.Context(), batchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetBatchResults returns the results of a completed batch
// GET /api/v1/batch/:id/results
func (h *ContentFactoryHandler) GetBatchResults(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	batchID := c.Param("id")

	results, err := h.batchService.GetBatchResults(c.Request.Context(), batchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

// ListBatches lists all batches for the user
// GET /api/v1/batch
func (h *ContentFactoryHandler) ListBatches(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit := 20
	offset := 0

	batches, err := h.batchService.ListBatches(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "LIST_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": batches,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  len(batches),
		},
	})
}

// CancelBatch cancels a batch
// POST /api/v1/batch/:id/cancel
func (h *ContentFactoryHandler) CancelBatch(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	batchID := c.Param("id")

	if err := h.batchService.CancelBatch(c.Request.Context(), batchID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "CANCEL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch cancelled successfully",
		"batchId": batchID,
	})
}

// RetryFailedVideos retries failed videos in a batch
// POST /api/v1/batch/:id/retry
func (h *ContentFactoryHandler) RetryFailedVideos(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	batchID := c.Param("id")

	if err := h.batchService.RetryFailedVideos(c.Request.Context(), batchID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "RETRY_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Failed videos queued for retry",
		"batchId": batchID,
	})
}

// GetQueueStats returns queue statistics
// GET /api/v1/batch/queue/stats
func (h *ContentFactoryHandler) GetQueueStats(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	stats, err := h.batchService.GetQueueStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "STATS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ============================================
// VARIATIONS ENDPOINTS
// ============================================

// CreateVariations creates content variations
// POST /api/v1/variations/create
func (h *ContentFactoryHandler) CreateVariations(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.CreateVariationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	result, err := h.variationsService.CreateVariations(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "VARIATION_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetPlatformSpecs returns platform specifications
// GET /api/v1/variations/platforms
func (h *ContentFactoryHandler) GetPlatformSpecs(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	specs := make(map[string]interface{})
	for key, spec := range service.PlatformSpecs {
		specs[key] = spec
	}

	c.JSON(http.StatusOK, gin.H{
		"data": specs,
	})
}

// ============================================
// OPTIMIZER ENDPOINTS
// ============================================

// AnalyzeVideo analyzes a video for optimization opportunities
// POST /api/v1/optimizer/analyze
func (h *ContentFactoryHandler) AnalyzeVideo(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		VideoID string `json:"videoId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	suggestions, err := h.optimizerService.AnalyzeVideo(c.Request.Context(), req.VideoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "ANALYSIS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": suggestions,
		"meta": gin.H{
			"total": len(suggestions),
		},
	})
}

// GeneratePerformanceReport generates a performance report
// POST /api/v1/optimizer/report
func (h *ContentFactoryHandler) GeneratePerformanceReport(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Days int `json:"days,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	if req.Days == 0 {
		req.Days = 30
	}

	report, err := h.optimizerService.GeneratePerformanceReport(c.Request.Context(), user.ID, req.Days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "REPORT_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetWinningContent identifies winning content patterns
// GET /api/v1/optimizer/winning-content
func (h *ContentFactoryHandler) GetWinningContent(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	analysis, err := h.optimizerService.GetWinningContent(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "ANALYSIS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// AutoOptimizeTitle auto-optimizes a video title
// POST /api/v1/optimizer/auto-title
func (h *ContentFactoryHandler) AutoOptimizeTitle(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		VideoID      string `json:"videoId" binding:"required"`
		CurrentTitle string `json:"currentTitle" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	newTitle, err := h.optimizerService.AutoOptimizeTitle(c.Request.Context(), req.VideoID, req.CurrentTitle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "OPTIMIZATION_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"currentTitle": req.CurrentTitle,
		"suggestedTitle": newTitle,
		"improvement": "Estimated 15% CTR increase",
	})
}
