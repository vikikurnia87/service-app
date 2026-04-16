package handler

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"

	"service-app/internal/dto"
	"service-app/internal/helpers"
	"service-app/internal/service"
	"service-app/internal/structs"
	"service-app/pkg/response"
)

// RoleHandler handles HTTP requests for role operations.
type RoleHandler struct {
	svc    service.RoleService
	logger *slog.Logger
}

// NewRoleHandler creates a new RoleHandler with injected dependencies.
func NewRoleHandler(svc service.RoleService, logger *slog.Logger) *RoleHandler {
	return &RoleHandler{
		svc:    svc,
		logger: logger,
	}
}

// GetAll handles GET /roles
func (h *RoleHandler) GetAll(c *echo.Context) error {
	ctx := c.Request().Context()

	pagination := helpers.GetPaginationParams(c, 15)
	orders := helpers.ParseOrderParamsWithDefault(c, structs.RoleOrderMapping, structs.RoleDefaultOrders)
	search := c.QueryParam("search_like")

	params := structs.ListParams{
		Pagination: pagination,
		Orders:     orders,
		Search:     search,
	}

	result, err := h.svc.GetAll(ctx, params)
	if err != nil {
		return response.Error(c, err)
	}

	return response.SuccessWithMeta(c, "roles retrieved successfully", result.Data, &result.Meta)
}

// GetByID handles GET /roles/:id
func (h *RoleHandler) GetByID(c *echo.Context) error {
	ctx := c.Request().Context()

	id, err := echo.PathParam[int64](c, "id")
	if err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, "invalid role id")
	}

	role, err := h.svc.GetByID(ctx, id)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, "role retrieved successfully", role)
}

// Create handles POST /roles
func (h *RoleHandler) Create(c *echo.Context) error {
	ctx := c.Request().Context()

	var req dto.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, "invalid request body")
	}

	if req.RoleName == "" || req.RoleCode == "" {
		return response.ErrorWithCode(c, http.StatusBadRequest, "role_name and role_code are required")
	}

	role, err := h.svc.Create(ctx, req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Created(c, "role created successfully", role)
}

// Update handles PUT /roles/:id
func (h *RoleHandler) Update(c *echo.Context) error {
	ctx := c.Request().Context()

	id, err := echo.PathParam[int64](c, "id")
	if err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, "invalid role id")
	}

	var req dto.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, "invalid request body")
	}

	role, err := h.svc.Update(ctx, id, req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, "role updated successfully", role)
}

// Delete handles DELETE /roles/:id
func (h *RoleHandler) Delete(c *echo.Context) error {
	ctx := c.Request().Context()

	id, err := echo.PathParam[int64](c, "id")
	if err != nil {
		return response.ErrorWithCode(c, http.StatusBadRequest, "invalid role id")
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, "role deleted successfully", nil)
}
