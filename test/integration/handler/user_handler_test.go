//go:build integration

package handler_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"service-app/internal/cache"
	"service-app/internal/handler"
	"service-app/internal/repository"
	"service-app/internal/service"
	appRedis "service-app/pkg/redis"
	"service-app/test/integration/testhelper"
)

// setupIntegrationUserEcho creates a full Echo stack backed by a real test DB.
func setupIntegrationUserEcho(t *testing.T) *echo.Echo {
	t.Helper()

	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	logger := slog.Default()

	var c cache.Cache
	redisCli, err := appRedis.NewRedisClient(testhelper.LoadRedisConfig(t))
	if err == nil {
		c = cache.NewRedisCache(redisCli)
		t.Cleanup(func() { redisCli.Close() })
	} else {
		c = cache.NewNoopCache()
	}

	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo, c, logger)
	h := handler.NewUserHandler(svc, logger)

	e := echo.New()
	e.POST("/users", h.Create)
	e.GET("/users", h.GetAll)
	e.GET("/users/:id", h.GetByID)
	e.PUT("/users/:id", h.Update)
	e.DELETE("/users/:id", h.Delete)
	return e
}

func TestUserHandler_Integration_CreateAndGet(t *testing.T) {
	e := setupIntegrationUserEcho(t)

	body := `{"name":"Int User","email":"intuser@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)

	req = httptest.NewRequest(http.MethodGet, "/users", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	resp = parseResp(t, rec)
	assert.True(t, resp.Success)
}

func TestUserHandler_Integration_GetByID_NotFound(t *testing.T) {
	e := setupIntegrationUserEcho(t)

	req := httptest.NewRequest(http.MethodGet, "/users/99999", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUserHandler_Integration_Delete(t *testing.T) {
	e := setupIntegrationUserEcho(t)

	body := `{"name":"Del User","email":"deluser@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	req = httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
