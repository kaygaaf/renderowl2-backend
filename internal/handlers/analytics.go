package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"renderowl-api/internal/middleware"
	"renderowl-api/internal/service"
)

// AnalyticsHandler handles analytics HTTP requests
type AnalyticsHandler struct {
	service *service.AnalyticsService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(service *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

// GetOverview returns the analytics overview
func (h *AnalyticsHandler) GetOverview(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get days parameter (default 30)
	days := 30
	if d := c.Query("days"); d != "" {
		if val, err := strconv.Atoi(d); err == nil && val > 0 {
			days = val
		}
	}

	overview, err := h.service.GetAnalyticsOverview(c.Request.Context(), user.ID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetDashboardSummary returns the main dashboard summary
func (h *AnalyticsHandler) GetDashboardSummary(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	summary, err := h.service.GetDashboardSummary(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetVideoPerformance returns video performance data
func (h *AnalyticsHandler) GetVideoPerformance(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Pagination
	limit := 20
	offset := 0
	
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	performance, err := h.service.GetVideoPerformance(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": performance.Videos,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  performance.TotalCount,
		},
	})
}

// GetPlatformBreakdown returns platform breakdown data
func (h *AnalyticsHandler) GetPlatformBreakdown(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get days parameter (default 30)
	days := 30
	if d := c.Query("days"); d != "" {
		if val, err := strconv.Atoi(d); err == nil && val > 0 {
			days = val
		}
	}

	breakdown, err := h.service.GetPlatformBreakdown(c.Request.Context(), user.ID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, breakdown)
}

// GetEngagementMetrics returns engagement metrics
func (h *AnalyticsHandler) GetEngagementMetrics(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get days parameter (default 30)
	days := 30
	if d := c.Query("days"); d != "" {
		if val, err := strconv.Atoi(d); err == nil && val > 0 {
			days = val
		}
	}

	metrics, err := h.service.GetEngagementMetrics(c.Request.Context(), user.ID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetUserGrowth returns user growth data
func (h *AnalyticsHandler) GetUserGrowth(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get days parameter (default 30)
	days := 30
	if d := c.Query("days"); d != "" {
		if val, err := strconv.Atoi(d); err == nil && val > 0 {
			days = val
		}
	}

	growth, err := h.service.GetUserGrowth(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, growth)
}

// TrackViewRequest represents a view tracking request
func (h *AnalyticsHandler) TrackView(c *gin.Context) {
	var req service.TrackViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	if err := h.service.TrackView(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "View tracked successfully"})
}

// TrackEngagementRequest represents engagement tracking request
func (h *AnalyticsHandler) TrackEngagement(c *gin.Context) {
	var req service.TrackEngagementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	if err := h.service.TrackEngagement(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Engagement tracked successfully"})
}

// ExportAnalytics exports analytics data
func (h *AnalyticsHandler) ExportAnalytics(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse date range
	startDate := time.Now().UTC().AddDate(0, 0, -30)
	endDate := time.Now().UTC()

	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = t
		}
	}
	if e := c.Query("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			endDate = t
		}
	}

	format := c.DefaultQuery("format", "json")

	req := &service.ExportAnalyticsRequest{
		UserID:    user.ID,
		StartDate: startDate,
		EndDate:   endDate,
		Format:    format,
	}

	data, err := h.service.ExportAnalytics(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	// Set appropriate headers
	if format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=analytics.csv")
	} else {
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=analytics.json")
	}

	c.Data(http.StatusOK, c.Writer.Header().Get("Content-Type"), data)
}

// ReceiveWebhook handles incoming webhooks from platforms
func (h *AnalyticsHandler) ReceiveWebhook(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Platform is required",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	// Determine event type from payload
	eventType := "unknown"
	if et, ok := payload["event_type"].(string); ok {
		eventType = et
	}

	// Get video ID from payload
	videoID := ""
	if vid, ok := payload["video_id"].(string); ok {
		videoID = vid
	}

	req := &service.WebhookEventRequest{
		Platform:  platform,
		EventType: eventType,
		VideoID:   videoID,
		Payload:   payload,
	}

	eventID, err := h.service.StoreWebhookEvent(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":  "Webhook received",
		"event_id": eventID,
	})
}
