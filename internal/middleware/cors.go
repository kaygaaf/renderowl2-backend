package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"renderowl-api/internal/config"
)

// CORS configures CORS middleware
func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow configured frontend URL
		allowedOrigins := []string{
			cfg.FrontendURL,
			"http://localhost:3000",
			"http://localhost:3001",
			"https://staging.renderowl.app",
			"https://renderowl.app",
		}

		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				allowed = true
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// If no origin match in production, still set for development
		if !allowed && cfg.Environment == "development" {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
