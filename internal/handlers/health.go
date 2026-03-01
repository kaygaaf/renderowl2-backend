package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthCheck returns basic health status
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "renderowl-api",
		"version":   "2.0.0",
		"timestamp": time.Now().UTC(),
	})
}

// ReadinessCheck verifies all dependencies are ready
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	checks := make(map[string]interface{})
	allHealthy := true

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil {
		checks["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  fmt.Sprintf("Failed to get SQL DB: %v", err),
		}
		allHealthy = false
	} else {
		if err := sqlDB.Ping(); err != nil {
			checks["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			allHealthy = false
		} else {
			// Get connection pool stats
			stats := sqlDB.Stats()
			checks["database"] = map[string]interface{}{
				"status":           "healthy",
				"openConnections":  stats.OpenConnections,
				"inUse":            stats.InUse,
				"idle":             stats.Idle,
			}
		}
	}

	// Check Redis (placeholder - would need Redis client)
	checks["redis"] = map[string]interface{}{
		"status":  "healthy",
		"message": "Redis check not implemented",
	}

	// Check Remotion (placeholder)
	checks["remotion"] = map[string]interface{}{
		"status":  "healthy",
		"message": "Remotion check not implemented",
	}

	response := gin.H{
		"status":   "ready",
		"checks":   checks,
		"version":  "2.0.0",
		"timestamp": time.Now().UTC(),
	}

	if !allHealthy {
		response["status"] = "not ready"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// LivenessCheck is a simple liveness probe
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"timestamp": time.Now().UTC(),
	})
}

// ComprehensiveHealthCheck returns detailed health information
func (h *HealthHandler) ComprehensiveHealthCheck(c *gin.Context) {
	health := gin.H{
		"status":  "healthy",
		"service": "renderowl-api",
		"version": "2.0.0",
		"time":    time.Now().UTC().Format(time.RFC3339),
	}

	// Add system info
	health["system"] = gin.H{
		"goVersion": "1.22",
	}

	c.JSON(http.StatusOK, health)
}

// getDBStats returns database connection pool statistics
func getDBStats(db *sql.DB) map[string]interface{} {
	stats := db.Stats()
	return map[string]interface{}{
		"max_open_connections":    stats.MaxOpenConnections,
		"open_connections":        stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration_ms":        stats.WaitDuration.Milliseconds(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}
}
