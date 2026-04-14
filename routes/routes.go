package routes

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"service-app/internal/handler"
	custommw "service-app/internal/middleware"
)

// Handlers groups all handler dependencies for route registration.
type Handlers struct {
	Health *handler.HealthHandler
	User   *handler.UserHandler
	Role   *handler.RoleHandler
}

// RegisterRoutes sets up all application routes and middleware.
func RegisterRoutes(e *echo.Echo, h Handlers, logger *slog.Logger) {
	// Global middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))
	e.Use(custommw.RequestLogger(logger))

	// Health check
	e.GET("/health", h.Health.Check)

	// API v1 group
	v1 := e.Group("/api/v1")

	// User routes
	users := v1.Group("/users")
	users.GET("", h.User.GetAll)
	users.GET("/:id", h.User.GetByID)
	users.POST("", h.User.Create)
	users.PUT("/:id", h.User.Update)
	users.DELETE("/:id", h.User.Delete)

	// Role routes
	roles := v1.Group("/roles")
	roles.GET("", h.Role.GetAll)
	roles.GET("/:id", h.Role.GetByID)
	roles.POST("", h.Role.Create)
	roles.PUT("/:id", h.Role.Update)
	roles.DELETE("/:id", h.Role.Delete)
}
