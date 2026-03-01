package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"renderowl-api/internal/middleware"
	"renderowl-api/internal/service"
)

// TimelineHandler handles timeline HTTP requests
type TimelineHandler struct {
	service *service.TimelineService
}

// NewTimelineHandler creates a new timeline handler
func NewTimelineHandler(service *service.TimelineService) *TimelineHandler {
	return &TimelineHandler{service: service}
}

// Create creates a new timeline
func (h *TimelineHandler) Create(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req service.CreateTimelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	timeline, err := h.service.Create(user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, timeline)
}

// Get retrieves a timeline by ID
func (h *TimelineHandler) Get(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")
	timeline, err := h.service.Get(id, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, timeline)
}

// List lists all timelines for the authenticated user
func (h *TimelineHandler) List(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit := 20
	offset := 0

	timelines, err := h.service.List(user.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": timelines,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  len(timelines),
		},
	})
}

// Update updates a timeline
func (h *TimelineHandler) Update(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	var req service.UpdateTimelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	timeline, err := h.service.Update(id, user.ID, &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, timeline)
}

// Delete deletes a timeline
func (h *TimelineHandler) Delete(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	if err := h.service.Delete(id, user.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
