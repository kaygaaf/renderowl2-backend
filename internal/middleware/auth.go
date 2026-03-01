package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"renderowl-api/internal/config"
	"renderowl-api/internal/domain"
)

const (
	UserContextKey = "user"
)

// Auth middleware validates Clerk JWT tokens
func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "AUTH_MISSING",
			})
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Use 'Bearer <token>'",
				"code":  "AUTH_INVALID_FORMAT",
			})
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			// In production, fetch Clerk's JWKS to get the public key
			// For now, use the secret key from config
			return []byte(cfg.ClerkSecretKey), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "AUTH_INVALID_TOKEN",
			})
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
				"code":  "AUTH_INVALID_CLAIMS",
			})
			return
		}

		// Extract user ID from sub claim
		userID, ok := claims["sub"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found in token",
				"code":  "AUTH_MISSING_USER_ID",
			})
			return
		}

		// Extract email if available
		email := ""
		if emailClaim, ok := claims["email"].(string); ok {
			email = emailClaim
		}

		// Set user context
		user := &domain.UserContext{
			ID:    userID,
			Email: email,
		}
		c.Set(UserContextKey, user)

		c.Next()
	}
}

// GetUser retrieves the authenticated user from context
func GetUser(c *gin.Context) *domain.UserContext {
	user, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}
	return user.(*domain.UserContext)
}
