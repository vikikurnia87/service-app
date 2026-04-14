//go:build integration

package handler

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
	"service-app/internal/repository"
	"service-app/internal/service"
	appRedis "service-app/pkg/redis"
	"service-app/test/integration/testhelper"
)

// setupIntegrationRoleEcho creates a full Echo stack backed by a real test DB.
func setupIntegrationRoleEcho(t *testing.T) *echo.Echo {
	t.Helper()

	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_role")
		testhelper.TeardownTestDB(t, db)
	})

	logger := slog.Default()

	var c cache.Cache
	redisCli, err := appRedis.NewRedisClient(testhelper.LoadRedisConfig(t))
	if err == nil {
		c = cache.NewRedisCache(redisCli)
		t.Cleanup(func() { redisCli.Close() })
	} else {
		c = cache.NewNoopCache()
	}

	repo := repository.NewRoleRepository(db)
	svc := service.NewRoleService(repo, c, logger)
	h := NewRoleHandler(svc, logger)

	e := echo.New()
	e.POST("/roles", h.Create)
	e.GET("/roles", h.GetAll)
	e.GET("/roles/:id", h.GetByID)
	e.PUT("/roles/:id", h.Update)
	e.DELETE("/roles/:id", h.Delete)
	return e
}

func TestRoleHandler_Integration_CreateAndGet(t *testing.T) {
	e := setupIntegrationRoleEcho(t)

	body := `{"role_name":"IntAdmin","role_desc":"Integration admin","role_code":"INT_ADMIN"}`
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	resp := parseResp(t, rec)
	assert.True(t, resp.Success)

	req = httptest.NewRequest(http.MethodGet, "/roles", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	resp = parseResp(t, rec)
	assert.True(t, resp.Success)
}

func TestRoleHandler_Integration_GetByID_NotFound(t *testing.T) {
	e := setupIntegrationRoleEcho(t)

	req := httptest.NewRequest(http.MethodGet, "/roles/99999", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRoleHandler_Integration_Delete(t *testing.T) {
	e := setupIntegrationRoleEcho(t)

	body := `{"role_name":"DelRole","role_desc":"to delete","role_code":"DEL_ROLE"}`
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	req = httptest.NewRequest(http.MethodDelete, "/roles/1", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
