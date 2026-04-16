package handler_test

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
	"service-app/internal/handler"
	"service-app/internal/structs"
	"service-app/pkg/apperror"
	appmock "service-app/test/mock"
)

// setupUserEcho creates a fresh Echo instance with user handler routes.
func setupUserEcho(t *testing.T, mockSvc *appmock.MockUserService) *echo.Echo {
	t.Helper()
	h := handler.NewUserHandler(mockSvc, slog.Default())

	e := echo.New()
	e.GET("/users", h.GetAll)
	e.GET("/users/:id", h.GetByID)
	e.POST("/users", h.Create)
	e.PUT("/users/:id", h.Update)
	e.DELETE("/users/:id", h.Delete)
	return e
}

// ──────────────────────────────────────────────────────────────────────────────
// GET /users
// ──────────────────────────────────────────────────────────────────────────────

func TestUserHandler_GetAll_Success(t *testing.T) {
	mockSvc := new(appmock.MockUserService)
	e := setupUserEcho(t, mockSvc)

	users := []dto.UserResponse{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
	}
	paginatedResp := &dto.PaginatedResponse{
		Data: users,
		Meta: structs.Meta{Count: 2, Total: 2, TotalPages: 1, PerPage: 15, CurrentPage: 1},
	}
	mockSvc.On("GetAll", mock.Anything, mock.Anything).Return(paginatedResp, nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)
	assert.Equal(t, "users retrieved successfully", resp.Message)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_GetAll_Error(t *testing.T) {
	mockSvc := new(appmock.MockUserService)
	e := setupUserEcho(t, mockSvc)

	mockSvc.On("GetAll", mock.Anything, mock.Anything).Return(nil, apperror.NewInternal(nil))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	resp := parseResp(t, rec)
	assert.False(t, resp.Success)
	mockSvc.AssertExpectations(t)
}

// ──────────────────────────────────────────────────────────────────────────────
// GET /users/:id
// ──────────────────────────────────────────────────────────────────────────────

func TestUserHandler_GetByID_Success(t *testing.T) {
	mockSvc := new(appmock.MockUserService)
	e := setupUserEcho(t, mockSvc)

	user := &dto.UserResponse{ID: 1, Name: "Alice", Email: "alice@example.com"}
	mockSvc.On("GetByID", mock.Anything, int64(1)).Return(user, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_GetByID_InvalidID(t *testing.T) {
	mockSvc := new(appmock.MockUserService)
	e := setupUserEcho(t, mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	resp := parseResp(t, rec)
	assert.False(t, resp.Success)
	assert.Equal(t, "invalid user id", resp.Message)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	mockSvc := new(appmock.MockUserService)
	e := setupUserEcho(t, mockSvc)

	mockSvc.On("GetByID", mock.Anything, int64(99)).
		Return(nil, apperror.NewNotFound("user with id 99 not found"))

	req := httptest.NewRequest(http.MethodGet, "/users/99", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /users
// ──────────────────────────────────────────────────────────────────────────────

func TestUserHandler_Create_Success(t *testing.T) {
	mockSvc := new(appmock.MockUserService)
	e := setupUserEcho(t, mockSvc)

	createReq := dto.CreateUserRequest{Name: "New", Email: "new@example.com"}
	mockSvc.On("Create", mock.Anything, createReq).
		Return(&dto.UserResponse{ID: 1, Name: "New", Email: "new@example.com"}, nil)

	body := `{"name":"New","email":"new@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)
	assert.Equal(t, "user created successfully", resp.Message)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Create_MissingFields(t *testing.T) {
	mockSvc := new(appmock.MockUserService)
	e := setupUserEcho(t, mockSvc)

	body := `{"name":""}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	resp := parseResp(t, rec)
	assert.Equal(t, "name and email are required", resp.Message)
}

// ──────────────────────────────────────────────────────────────────────────────
// PUT /users/:id
// ──────────────────────────────────────────────────────────────────────────────

func TestUserHandler_Update_Success(t *testing.T) {
	mockSvc := new(appmock.MockUserService)
	e := setupUserEcho(t, mockSvc)

	updateReq := dto.UpdateUserRequest{Name: "Updated"}
	mockSvc.On("Update", mock.Anything, int64(1), updateReq).
		Return(&dto.UserResponse{ID: 1, Name: "Updated", Email: "o@example.com"}, nil)

	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/users/1", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(appmock.MockUserService)
	e := setupUserEcho(t, mockSvc)

	req := httptest.NewRequest(http.MethodPut, "/users/abc", strings.NewReader(`{"name":"X"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ──────────────────────────────────────────────────────────────────────────────
// DELETE /users/:id
// ──────────────────────────────────────────────────────────────────────────────

func TestUserHandler_Delete_Success(t *testing.T) {
	mockSvc := new(appmock.MockUserService)
	e := setupUserEcho(t, mockSvc)

	mockSvc.On("Delete", mock.Anything, int64(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(appmock.MockUserService)
	e := setupUserEcho(t, mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/users/abc", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
