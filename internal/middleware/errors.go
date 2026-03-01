package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// ErrorHandler middleware handles errors consistently
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			var response ErrorResponse
			switch {
			case errors.Is(err.Err, validator.ValidationErrors{}):
				response = ErrorResponse{
					Error: "Validation failed",
					Code:  "VALIDATION_ERROR",
					Details: validationErrorsToMap(err.Err.(validator.ValidationErrors)),
				}
				c.JSON(http.StatusBadRequest, response)

			case errors.Is(err.Err, ErrNotFound):
				response = ErrorResponse{
					Error: err.Err.Error(),
					Code:  "NOT_FOUND",
				}
				c.JSON(http.StatusNotFound, response)

			case errors.Is(err.Err, ErrForbidden):
				response = ErrorResponse{
					Error: err.Err.Error(),
					Code:  "FORBIDDEN",
				}
				c.JSON(http.StatusForbidden, response)

			case errors.Is(err.Err, ErrConflict):
				response = ErrorResponse{
					Error: err.Err.Error(),
					Code:  "CONFLICT",
				}
				c.JSON(http.StatusConflict, response)

			default:
				response = ErrorResponse{
					Error: "Internal server error",
					Code:  "INTERNAL_ERROR",
				}
				c.JSON(http.StatusInternalServerError, response)
			}
		}
	}
}

// Custom error types
var (
	ErrNotFound  = errors.New("resource not found")
	ErrForbidden = errors.New("access forbidden")
	ErrConflict  = errors.New("resource conflict")
)

func validationErrorsToMap(errs validator.ValidationErrors) map[string]string {
	details := make(map[string]string)
	for _, err := range errs {
		details[err.Field()] = err.Tag()
	}
	return details
}
