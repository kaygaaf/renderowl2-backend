package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"renderowl-api/internal/middleware"
	"renderowl-api/internal/service"
)

// ClipHandler handles clip HTTP requests
type ClipHandler struct {
	service *service.ClipService
}

// NewClipHandler creates a new clip handler
func NewClipHandler(service *service.ClipService) *ClipHandler {
	return &ClipHandler{service: service}
}

// Create creates a new clip
func (h *ClipHandler) Create(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	timelineID := c.Param("id")

	var req service.CreateClipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	clip, err := h.service.Create(user.ID, timelineID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "BAD_REQUEST",
		})
		return
	}

	c.JSON(http.StatusCreated, clip)
}

// Get retrieves a clip by ID
func (h *ClipHandler) Get(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	clipID := c.Param("clipId")

	clip, err := h.service.Get(user.ID, clipID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, clip)
}

// List lists all clips for a timeline
func (h *ClipHandler) List(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	timelineID := c.Param("id")

	clips, err := h.service.ListByTimeline(user.ID, timelineID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "BAD_REQUEST",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": clips,
		"meta": gin.H{
			"timelineId": timelineID,
			"total":      len(clips),
		},
	})
}

// Update updates a clip
func (h *ClipHandler) Update(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	clipID := c.Param("clipId")

	var req service.UpdateClipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	clip, err := h.service.Update(user.ID, clipID, &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, clip)
}

// Delete deletes a clip
func (h *ClipHandler) Delete(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	clipID := c.Param("clipId")

	if err := h.service.Delete(user.ID, clipID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
