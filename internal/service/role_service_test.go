package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"service-app/internal/dto"
	"service-app/internal/mocks"
	"service-app/internal/model"
	"service-app/internal/structs"
)

// newTestRoleService creates a RoleService with mocked dependencies.
func newTestRoleService(t *testing.T) (RoleService, *mocks.MockRoleRepository, *mocks.MockCache) {
	t.Helper()
	mockRepo := new(mocks.MockRoleRepository)
	mockCache := new(mocks.MockCache)
	logger := slog.Default()
	svc := NewRoleService(mockRepo, mockCache, logger)
	return svc, mockRepo, mockCache
}

// ──────────────────────────────────────────────────────────────────────────────
// GetAll
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleService_GetAll_Success(t *testing.T) {
	svc, mockRepo, mockCache := newTestRoleService(t)
	ctx := context.Background()

	params := structs.ListParams{
		Pagination: structs.Pagination{Page: 1, Limit: 15, Offset: 0},
		Orders:     structs.RoleDefaultOrders,
	}

	roles := []model.Role{
		{ID: 1, RoleName: "Admin", RoleDesc: "Administrator", RoleCode: "ADMIN", Status: 1},
		{ID: 2, RoleName: "User", RoleDesc: "User", RoleCode: "USER", Status: 1},
	}

	mockCache.On("Get", ctx, mock.Anything, mock.Anything).Return(errors.New("miss"))
	mockRepo.On("FindPaginated", ctx, params).Return(roles, 2, nil)
	mockCache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	result, err := svc.GetAll(ctx, params)

	require.NoError(t, err)
	data := result.Data.([]dto.RoleResponse)
	assert.Len(t, data, 2)
	assert.Equal(t, "Admin", data[0].RoleName)
	assert.Equal(t, 2, result.Meta.Total)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_GetAll_Error(t *testing.T) {
	svc, mockRepo, mockCache := newTestRoleService(t)
	ctx := context.Background()

	params := structs.ListParams{
		Pagination: structs.Pagination{Page: 1, Limit: 15, Offset: 0},
		Orders:     structs.RoleDefaultOrders,
	}

	mockCache.On("Get", ctx, mock.Anything, mock.Anything).Return(errors.New("miss"))
	mockRepo.On("FindPaginated", ctx, params).Return(nil, 0, errors.New("db error"))

	result, err := svc.GetAll(ctx, params)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ──────────────────────────────────────────────────────────────────────────────
// GetByID
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleService_GetByID_CacheHit(t *testing.T) {
	svc, mockRepo, mockCache := newTestRoleService(t)
	ctx := context.Background()

	mockCache.On("Get", ctx, "role:1", mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(2).(*dto.RoleResponse)
			dest.ID = 1
			dest.RoleName = "Cached Role"
		}).
		Return(nil)

	result, err := svc.GetByID(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, "Cached Role", result.RoleName)
	mockRepo.AssertNotCalled(t, "FindByID")
}

func TestRoleService_GetByID_CacheMiss(t *testing.T) {
	svc, mockRepo, mockCache := newTestRoleService(t)
	ctx := context.Background()

	mockCache.On("Get", ctx, "role:1", mock.Anything).Return(errors.New("miss"))

	role := &model.Role{ID: 1, RoleName: "Admin", RoleDesc: "desc", RoleCode: "ADMIN", Status: 1}
	mockRepo.On("FindByID", ctx, int64(1)).Return(role, nil)
	mockCache.On("Set", ctx, "role:1", mock.Anything, mock.Anything).Return(nil)

	result, err := svc.GetByID(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, "Admin", result.RoleName)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo, mockCache := newTestRoleService(t)
	ctx := context.Background()

	mockCache.On("Get", ctx, "role:99", mock.Anything).Return(errors.New("miss"))
	mockRepo.On("FindByID", ctx, int64(99)).Return(nil, sql.ErrNoRows)

	result, err := svc.GetByID(ctx, 99)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

// ──────────────────────────────────────────────────────────────────────────────
// Create
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleService_Create_Success(t *testing.T) {
	svc, mockRepo, mockCache := newTestRoleService(t)
	ctx := context.Background()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Role")).
		Run(func(args mock.Arguments) {
			r := args.Get(1).(*model.Role)
			r.ID = 1
		}).
		Return(nil)
	mockCache.On("DeleteByPrefix", ctx, roleListCachePrefix).Return(nil)

	req := dto.CreateRoleRequest{RoleName: "Admin", RoleDesc: "desc", RoleCode: "ADMIN"}
	result, err := svc.Create(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Admin", result.RoleName)
}

func TestRoleService_Create_Error(t *testing.T) {
	svc, mockRepo, _ := newTestRoleService(t)
	ctx := context.Background()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Role")).Return(errors.New("dup"))

	req := dto.CreateRoleRequest{RoleName: "Dup", RoleDesc: "dup", RoleCode: "DUP"}
	result, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ──────────────────────────────────────────────────────────────────────────────
// Update
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleService_Update_Success(t *testing.T) {
	svc, mockRepo, mockCache := newTestRoleService(t)
	ctx := context.Background()

	existing := &model.Role{ID: 1, RoleName: "Old", RoleDesc: "old", RoleCode: "OLD", Status: 1}
	mockRepo.On("FindByID", ctx, int64(1)).Return(existing, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Role")).Return(nil)
	mockCache.On("Delete", ctx, "role:1").Return(nil)
	mockCache.On("DeleteByPrefix", ctx, roleListCachePrefix).Return(nil)

	req := dto.UpdateRoleRequest{RoleName: "New"}
	result, err := svc.Update(ctx, 1, req)

	require.NoError(t, err)
	assert.Equal(t, "New", result.RoleName)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestRoleService_Update_NotFound(t *testing.T) {
	svc, mockRepo, _ := newTestRoleService(t)
	ctx := context.Background()

	mockRepo.On("FindByID", ctx, int64(99)).Return(nil, sql.ErrNoRows)

	req := dto.UpdateRoleRequest{RoleName: "X"}
	result, err := svc.Update(ctx, 99, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ──────────────────────────────────────────────────────────────────────────────
// Delete
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleService_Delete_Success(t *testing.T) {
	svc, mockRepo, mockCache := newTestRoleService(t)
	ctx := context.Background()

	mockRepo.On("Delete", ctx, int64(1)).Return(nil)
	mockCache.On("Delete", ctx, "role:1").Return(nil)
	mockCache.On("DeleteByPrefix", ctx, roleListCachePrefix).Return(nil)

	err := svc.Delete(ctx, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestRoleService_Delete_Error(t *testing.T) {
	svc, mockRepo, _ := newTestRoleService(t)
	ctx := context.Background()

	mockRepo.On("Delete", ctx, int64(1)).Return(errors.New("db error"))

	err := svc.Delete(ctx, 1)

	assert.Error(t, err)
}
