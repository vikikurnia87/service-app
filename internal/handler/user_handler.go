package handler

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"

	"service-app/internal/dto"
	"service-app/internal/service"
	"service-app/pkg/response"
)

// UserHandler handles HTTP requests for user operations.
type UserHandler struct {
	svc    service.UserService
	logger *slog.Logger
}

// NewUserHandler creates a new UserHandler with injected dependencies.
func NewUserHandler(svc service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		svc:    svc,
		logger: logger,
	}
}

// GetAll handles GET /users
func (h *UserHandler) GetAll(c *echo.Context) error {
	ctx := c.Request().Context()

	users, err := h.svc.GetAll(ctx)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, "users retrieved successfully", users)
}

// GetByID handles GET /users/:id
func (h *UserHandler) GetByID(c *echo.Context) error {
	ctx := c.Request().Context()

	id, err := echo.PathParam[int64](c, "id")
	if err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, "invalid user id")
	}

	user, err := h.svc.GetByID(ctx, id)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, "user retrieved successfully", user)
}

// Create handles POST /users
func (h *UserHandler) Create(c *echo.Context) error {
	ctx := c.Request().Context()

	var req dto.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" || req.Email == "" {
		return response.ErrorWithCode(c, http.StatusBadRequest, "name and email are required")
	}

	user, err := h.svc.Create(ctx, req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Created(c, "user created successfully", user)
}

// Update handles PUT /users/:id
func (h *UserHandler) Update(c *echo.Context) error {
	ctx := c.Request().Context()

	id, err := echo.PathParam[int64](c, "id")
	if err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, "invalid user id")
	}

	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, "invalid request body")
	}

	user, err := h.svc.Update(ctx, id, req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, "user updated successfully", user)
}

// Delete handles DELETE /users/:id
func (h *UserHandler) Delete(c *echo.Context) error {
	ctx := c.Request().Context()

	id, err := echo.PathParam[int64](c, "id")
	if err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, "invalid user id")
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, "user deleted successfully", nil)
}
