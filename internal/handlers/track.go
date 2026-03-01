package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"renderowl-api/internal/middleware"
	"renderowl-api/internal/service"
)

// TrackHandler handles track HTTP requests
type TrackHandler struct {
	service *service.TrackService
}

// NewTrackHandler creates a new track handler
func NewTrackHandler(service *service.TrackService) *TrackHandler {
	return &TrackHandler{service: service}
}

// Create creates a new track
func (h *TrackHandler) Create(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	timelineID := c.Param("id")

	var req service.CreateTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	track, err := h.service.Create(user.ID, timelineID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "BAD_REQUEST",
		})
		return
	}

	c.JSON(http.StatusCreated, track)
}

// Get retrieves a track by ID
func (h *TrackHandler) Get(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trackID := c.Param("trackId")

	track, err := h.service.Get(user.ID, trackID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, track)
}

// List lists all tracks for a timeline
func (h *TrackHandler) List(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	timelineID := c.Param("id")

	tracks, err := h.service.ListByTimeline(user.ID, timelineID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "BAD_REQUEST",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tracks,
		"meta": gin.H{
			"timelineId": timelineID,
			"total":      len(tracks),
		},
	})
}

// Update updates a track
func (h *TrackHandler) Update(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trackID := c.Param("trackId")

	var req service.UpdateTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	track, err := h.service.Update(user.ID, trackID, &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, track)
}

// Delete deletes a track
func (h *TrackHandler) Delete(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trackID := c.Param("trackId")

	if err := h.service.Delete(user.ID, trackID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Reorder reorders tracks
func (h *TrackHandler) Reorder(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	timelineID := c.Param("id")

	var req service.ReorderTracksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	if err := h.service.Reorder(user.ID, timelineID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "BAD_REQUEST",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tracks reordered successfully"})
}

// ToggleMute toggles track mute status
func (h *TrackHandler) ToggleMute(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trackID := c.Param("trackId")

	track, err := h.service.ToggleMute(user.ID, trackID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, track)
}

// ToggleSolo toggles track solo status
func (h *TrackHandler) ToggleSolo(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trackID := c.Param("trackId")

	track, err := h.service.ToggleSolo(user.ID, trackID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, track)
}
