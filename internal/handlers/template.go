package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"renderowl-api/internal/domain"
	"renderowl-api/internal/middleware"
	"renderowl-api/internal/service"
)

// TemplateHandler handles template HTTP requests
type TemplateHandler struct {
	service *service.TemplateService
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(service *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{service: service}
}

// List lists all templates with optional filtering
func (h *TemplateHandler) List(c *gin.Context) {
	filter := domain.TemplateFilter{
		Category: c.Query("category"),
		Search:   c.Query("search"),
	}

	if limit, err := strconv.Atoi(c.Query("limit")); err == nil {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(c.Query("offset")); err == nil {
		filter.Offset = offset
	}

	templates, err := h.service.ListTemplates(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": templates,
		"meta": gin.H{
			"limit":  filter.Limit,
			"offset": filter.Offset,
			"total":  len(templates),
		},
	})
}

// Get retrieves a single template by ID
func (h *TemplateHandler) Get(c *gin.Context) {
	id := c.Param("id")

	template, err := h.service.GetTemplate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, template)
}

// GetCategories retrieves all template categories
func (h *TemplateHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categories,
	})
}

// Use creates a new timeline from a template
func (h *TemplateHandler) Use(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	templateID := c.Param("id")

	var req domain.UseTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	response, err := h.service.UseTemplate(templateID, user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetStats retrieves template statistics
func (h *TemplateHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetTemplateStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
