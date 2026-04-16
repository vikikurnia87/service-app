//go:build integration

package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"service-app/internal/cache"
	"service-app/internal/dto"
	"service-app/internal/repository"
	"service-app/internal/structs"
	appRedis "service-app/pkg/redis"
	"service-app/test/integration/testhelper"
)

// newIntegrationUserService creates a real service backed by test DB and cache (Redis or no-op).
func newIntegrationUserService(t *testing.T) UserService {
	t.Helper()
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_user")
		testhelper.TeardownTestDB(t, db)
	})

	repo := repository.NewUserRepository(db)

	// Use Redis if available, otherwise no-op
	var c cache.Cache
	redisCli, err := appRedis.NewRedisClient(testhelper.LoadRedisConfig(t))
	if err == nil {
		c = cache.NewRedisCache(redisCli)
		t.Cleanup(func() { redisCli.Close() })
	} else {
		c = cache.NewNoopCache()
	}

	logger := slog.Default()
	return NewUserService(repo, c, logger)
}

func TestUserService_Integration_CreateAndGetByID(t *testing.T) {
	svc := newIntegrationUserService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, dto.CreateUserRequest{
		Name:  "Integration User",
		Email: "int@example.com",
	})
	require.NoError(t, err)
	assert.NotZero(t, created.ID)

	found, err := svc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Integration User", found.Name)
}

func TestUserService_Integration_GetAll(t *testing.T) {
	svc := newIntegrationUserService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, dto.CreateUserRequest{Name: "A", Email: "a-int@example.com"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.CreateUserRequest{Name: "B", Email: "b-int@example.com"})
	require.NoError(t, err)

	params := structs.ListParams{
		Pagination: structs.Pagination{Page: 1, Limit: 15, Offset: 0},
		Orders:     structs.UserDefaultOrders,
	}

	result, err := svc.GetAll(ctx, params)
	require.NoError(t, err)
	data := result.Data.([]dto.UserResponse)
	assert.GreaterOrEqual(t, len(data), 2)
}

func TestUserService_Integration_Update(t *testing.T) {
	svc := newIntegrationUserService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, dto.CreateUserRequest{Name: "Old", Email: "old-int@example.com"})
	require.NoError(t, err)

	updated, err := svc.Update(ctx, created.ID, dto.UpdateUserRequest{Name: "New"})
	require.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
}

func TestUserService_Integration_Delete(t *testing.T) {
	svc := newIntegrationUserService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, dto.CreateUserRequest{Name: "Del", Email: "del-int@example.com"})
	require.NoError(t, err)

	err = svc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(ctx, created.ID)
	assert.Error(t, err)
}

func TestUserService_Integration_GetByID_NotFound(t *testing.T) {
	svc := newIntegrationUserService(t)
	ctx := context.Background()

	_, err := svc.GetByID(ctx, 99999)
	assert.Error(t, err)
}
