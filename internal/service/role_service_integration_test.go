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

// newIntegrationRoleService creates a real service backed by test DB.
func newIntegrationRoleService(t *testing.T) RoleService {
	t.Helper()
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_role")
		testhelper.TeardownTestDB(t, db)
	})

	repo := repository.NewRoleRepository(db)

	var c cache.Cache
	redisCli, err := appRedis.NewRedisClient(testhelper.LoadRedisConfig(t))
	if err == nil {
		c = cache.NewRedisCache(redisCli)
		t.Cleanup(func() { redisCli.Close() })
	} else {
		c = cache.NewNoopCache()
	}

	logger := slog.Default()
	return NewRoleService(repo, c, logger)
}

func TestRoleService_Integration_CreateAndGetByID(t *testing.T) {
	svc := newIntegrationRoleService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, dto.CreateRoleRequest{
		RoleName: "Admin",
		RoleDesc: "Administrator",
		RoleCode: "ADMIN_INT",
	})
	require.NoError(t, err)
	assert.NotZero(t, created.ID)

	found, err := svc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Admin", found.RoleName)
}

func TestRoleService_Integration_GetAll(t *testing.T) {
	svc := newIntegrationRoleService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, dto.CreateRoleRequest{RoleName: "A", RoleDesc: "a", RoleCode: "A_INT"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.CreateRoleRequest{RoleName: "B", RoleDesc: "b", RoleCode: "B_INT"})
	require.NoError(t, err)

	params := structs.ListParams{
		Pagination: structs.Pagination{Page: 1, Limit: 15, Offset: 0},
		Orders:     structs.RoleDefaultOrders,
	}

	result, err := svc.GetAll(ctx, params)
	require.NoError(t, err)
	data := result.Data.([]dto.RoleResponse)
	assert.GreaterOrEqual(t, len(data), 2)
}

func TestRoleService_Integration_Update(t *testing.T) {
	svc := newIntegrationRoleService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, dto.CreateRoleRequest{RoleName: "Old", RoleDesc: "old", RoleCode: "OLD_INT"})
	require.NoError(t, err)

	updated, err := svc.Update(ctx, created.ID, dto.UpdateRoleRequest{RoleName: "New"})
	require.NoError(t, err)
	assert.Equal(t, "New", updated.RoleName)
}

func TestRoleService_Integration_Delete(t *testing.T) {
	svc := newIntegrationRoleService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, dto.CreateRoleRequest{RoleName: "Del", RoleDesc: "del", RoleCode: "DEL_INT"})
	require.NoError(t, err)

	err = svc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(ctx, created.ID)
	assert.Error(t, err)
}
