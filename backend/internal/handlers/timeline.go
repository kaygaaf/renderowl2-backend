package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kaygaaf/renderowl2/internal/domain"
	"github.com/kaygaaf/renderowl2/internal/service"
)

// TimelineHandler handles timeline HTTP requests
type TimelineHandler struct {
	service service.TimelineService
}

// NewTimelineHandler creates a new timeline handler
func NewTimelineHandler(service service.TimelineService) *TimelineHandler {
	return &TimelineHandler{service: service}
}

// CreateTimelineRequest represents the request body for creating a timeline
// @Summary Create a new timeline
// @Description Create a new timeline for the authenticated user
// @Tags timelines
// @Accept json
// @Produce json
// @Param timeline body domain.CreateTimelineRequest true "Timeline data"
// @Success 201 {object} domain.TimelineResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/timeline [post]
func (h *TimelineHandler) CreateTimeline(c *gin.Context) {
	var req domain.CreateTimelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get userID from authenticated context
	userID := uint(1) // Placeholder - should come from JWT/auth middleware

	timeline, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, service.ToResponse(timeline))
}

// GetTimeline retrieves a timeline by ID
// @Summary Get a timeline
// @Description Get a timeline by its ID
// @Tags timelines
// @Produce json
// @Param id path int true "Timeline ID"
// @Success 200 {object} domain.TimelineResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/timeline/{id} [get]
func (h *TimelineHandler) GetTimeline(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timeline ID"})
		return
	}

	timeline, err := h.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "timeline not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "timeline not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, service.ToResponse(timeline))
}

// UpdateTimeline updates a timeline
// @Summary Update a timeline
// @Description Update an existing timeline
// @Tags timelines
// @Accept json
// @Produce json
// @Param id path int true "Timeline ID"
// @Param timeline body domain.UpdateTimelineRequest true "Timeline update data"
// @Success 200 {object} domain.TimelineResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/timeline/{id} [put]
func (h *TimelineHandler) UpdateTimeline(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timeline ID"})
		return
	}

	var req domain.UpdateTimelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get userID from authenticated context
	userID := uint(1) // Placeholder - should come from JWT/auth middleware

	timeline, err := h.service.Update(c.Request.Context(), uint(id), userID, req)
	if err != nil {
		switch err.Error() {
		case "timeline not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "timeline not found"})
		case "unauthorized: not your timeline":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, service.ToResponse(timeline))
}

// DeleteTimeline deletes a timeline
// @Summary Delete a timeline
// @Description Delete a timeline by its ID
// @Tags timelines
// @Param id path int true "Timeline ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/timeline/{id} [delete]
func (h *TimelineHandler) DeleteTimeline(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timeline ID"})
		return
	}

	// TODO: Get userID from authenticated context
	userID := uint(1) // Placeholder - should come from JWT/auth middleware

	if err := h.service.Delete(c.Request.Context(), uint(id), userID); err != nil {
		switch err.Error() {
		case "timeline not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "timeline not found"})
		case "unauthorized: not your timeline":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ListTimelines lists all timelines for the authenticated user
// @Summary List timelines
// @Description Get all timelines for the authenticated user
// @Tags timelines
// @Produce json
// @Param limit query int false "Limit (default 20)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {array} domain.TimelineResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/timelines [get]
func (h *TimelineHandler) ListTimelines(c *gin.Context) {
	// Parse pagination params
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// TODO: Get userID from authenticated context and filter by user
	timelines, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, service.ToResponseList(timelines))
}

// GetUserTimelines gets all timelines for a specific user
// @Summary Get user timelines
// @Description Get all timelines for a specific user
// @Tags timelines
// @Produce json
// @Success 200 {array} domain.TimelineResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/timelines/me [get]
func (h *TimelineHandler) GetUserTimelines(c *gin.Context) {
	// TODO: Get userID from authenticated context
	userID := uint(1) // Placeholder - should come from JWT/auth middleware

	timelines, err := h.service.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, service.ToResponseList(timelines))
}
