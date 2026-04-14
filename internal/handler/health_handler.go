package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// HealthHandler handles health check requests.
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check returns the health status of the application.
func (h *HealthHandler) Check(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "service-app",
	})
}
