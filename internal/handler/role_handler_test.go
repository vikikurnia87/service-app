package handler

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"service-app/internal/dto"
	"service-app/internal/mocks"
	"service-app/pkg/apperror"
)

// setupRoleEcho creates a fresh Echo instance with role handler routes.
func setupRoleEcho(t *testing.T, mockSvc *mocks.MockRoleService) *echo.Echo {
	t.Helper()
	h := NewRoleHandler(mockSvc, slog.Default())

	e := echo.New()
	e.GET("/roles", h.GetAll)
	e.GET("/roles/:id", h.GetByID)
	e.POST("/roles", h.Create)
	e.PUT("/roles/:id", h.Update)
	e.DELETE("/roles/:id", h.Delete)
	return e
}

// ──────────────────────────────────────────────────────────────────────────────
// GET /roles
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleHandler_GetAll_Success(t *testing.T) {
	mockSvc := new(mocks.MockRoleService)
	e := setupRoleEcho(t, mockSvc)

	roles := []dto.RoleResponse{
		{ID: 1, RoleName: "Admin", RoleDesc: "Administrator", RoleCode: "ADMIN", Status: 1},
		{ID: 2, RoleName: "User", RoleDesc: "User", RoleCode: "USER", Status: 1},
	}
	mockSvc.On("GetAll", mock.Anything).Return(roles, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)
	assert.Equal(t, "roles retrieved successfully", resp.Message)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_GetAll_Error(t *testing.T) {
	mockSvc := new(mocks.MockRoleService)
	e := setupRoleEcho(t, mockSvc)

	mockSvc.On("GetAll", mock.Anything).Return(nil, apperror.NewInternal(nil))

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

// ──────────────────────────────────────────────────────────────────────────────
// GET /roles/:id
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleHandler_GetByID_Success(t *testing.T) {
	mockSvc := new(mocks.MockRoleService)
	e := setupRoleEcho(t, mockSvc)

	role := &dto.RoleResponse{ID: 1, RoleName: "Admin", RoleCode: "ADMIN", Status: 1}
	mockSvc.On("GetByID", mock.Anything, int64(1)).Return(role, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles/1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_GetByID_InvalidID(t *testing.T) {
	mockSvc := new(mocks.MockRoleService)
	e := setupRoleEcho(t, mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/roles/abc", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	resp := parseResp(t, rec)
	assert.Equal(t, "invalid role id", resp.Message)
}

func TestRoleHandler_GetByID_NotFound(t *testing.T) {
	mockSvc := new(mocks.MockRoleService)
	e := setupRoleEcho(t, mockSvc)

	mockSvc.On("GetByID", mock.Anything, int64(99)).
		Return(nil, apperror.NewNotFound("role with id 99 not found"))

	req := httptest.NewRequest(http.MethodGet, "/roles/99", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /roles
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleHandler_Create_Success(t *testing.T) {
	mockSvc := new(mocks.MockRoleService)
	e := setupRoleEcho(t, mockSvc)

	createReq := dto.CreateRoleRequest{RoleName: "Editor", RoleDesc: "Content editor", RoleCode: "EDITOR"}
	mockSvc.On("Create", mock.Anything, createReq).
		Return(&dto.RoleResponse{ID: 1, RoleName: "Editor", RoleDesc: "Content editor", RoleCode: "EDITOR", Status: 1}, nil)

	body := `{"role_name":"Editor","role_desc":"Content editor","role_code":"EDITOR"}`
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)
	assert.Equal(t, "role created successfully", resp.Message)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_Create_MissingFields(t *testing.T) {
	mockSvc := new(mocks.MockRoleService)
	e := setupRoleEcho(t, mockSvc)

	body := `{"role_name":"","role_code":""}`
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	resp := parseResp(t, rec)
	assert.Equal(t, "role_name and role_code are required", resp.Message)
}

// ──────────────────────────────────────────────────────────────────────────────
// PUT /roles/:id
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleHandler_Update_Success(t *testing.T) {
	mockSvc := new(mocks.MockRoleService)
	e := setupRoleEcho(t, mockSvc)

	updateReq := dto.UpdateRoleRequest{RoleName: "Updated"}
	mockSvc.On("Update", mock.Anything, int64(1), updateReq).
		Return(&dto.RoleResponse{ID: 1, RoleName: "Updated", RoleCode: "ADMIN", Status: 1}, nil)

	body := `{"role_name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/roles/1", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(mocks.MockRoleService)
	e := setupRoleEcho(t, mockSvc)

	req := httptest.NewRequest(http.MethodPut, "/roles/abc", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ──────────────────────────────────────────────────────────────────────────────
// DELETE /roles/:id
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleHandler_Delete_Success(t *testing.T) {
	mockSvc := new(mocks.MockRoleService)
	e := setupRoleEcho(t, mockSvc)

	mockSvc.On("Delete", mock.Anything, int64(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/roles/1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(mocks.MockRoleService)
	e := setupRoleEcho(t, mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/roles/abc", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
