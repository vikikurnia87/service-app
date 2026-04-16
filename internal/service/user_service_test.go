package service

import (
	"context"
	"database/sql"
	"errors"
	"io"
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

// newTestUserService creates a UserService with mocked dependencies.
func newTestUserService(t *testing.T) (UserService, *mocks.MockUserRepository, *mocks.MockCache) {
	t.Helper()
	mockRepo := new(mocks.MockUserRepository)
	mockCache := new(mocks.MockCache)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewUserService(mockRepo, mockCache, logger)
	return svc, mockRepo, mockCache
}

// ──────────────────────────────────────────────────────────────────────────────
// GetAll
// ──────────────────────────────────────────────────────────────────────────────

func TestUserService_GetAll_Success(t *testing.T) {
	svc, mockRepo, mockCache := newTestUserService(t)
	ctx := context.Background()

	params := structs.ListParams{
		Pagination: structs.Pagination{Page: 1, Limit: 15, Offset: 0},
		Orders:     structs.UserDefaultOrders,
	}

	users := []model.User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}
	users[0].ID = 1
	users[1].ID = 2

	mockCache.On("Get", ctx, mock.Anything, mock.Anything).Return(errors.New("miss"))
	mockRepo.On("FindPaginated", ctx, params).Return(users, 2, nil)
	mockCache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	result, err := svc.GetAll(ctx, params)

	require.NoError(t, err)
	data := result.Data.([]dto.UserResponse)
	assert.Len(t, data, 2)
	assert.Equal(t, "Alice", data[0].Name)
	assert.Equal(t, "Bob", data[1].Name)
	assert.Equal(t, 2, result.Meta.Total)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetAll_Error(t *testing.T) {
	svc, mockRepo, mockCache := newTestUserService(t)
	ctx := context.Background()

	params := structs.ListParams{
		Pagination: structs.Pagination{Page: 1, Limit: 15, Offset: 0},
		Orders:     structs.UserDefaultOrders,
	}

	mockCache.On("Get", ctx, mock.Anything, mock.Anything).Return(errors.New("miss"))
	mockRepo.On("FindPaginated", ctx, params).Return(nil, 0, errors.New("db error"))

	result, err := svc.GetAll(ctx, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// ──────────────────────────────────────────────────────────────────────────────
// GetByID
// ──────────────────────────────────────────────────────────────────────────────

func TestUserService_GetByID_CacheHit(t *testing.T) {
	svc, mockRepo, mockCache := newTestUserService(t)
	ctx := context.Background()

	mockCache.On("Get", ctx, "user:1", mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(2).(*dto.UserResponse)
			dest.ID = 1
			dest.Name = "Cached"
			dest.Email = "cached@example.com"
		}).
		Return(nil)

	result, err := svc.GetByID(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, "Cached", result.Name)
	mockRepo.AssertNotCalled(t, "FindByID")
	mockCache.AssertExpectations(t)
}

func TestUserService_GetByID_CacheMiss(t *testing.T) {
	svc, mockRepo, mockCache := newTestUserService(t)
	ctx := context.Background()

	mockCache.On("Get", ctx, "user:1", mock.Anything).Return(errors.New("miss"))

	user := &model.User{Name: "DB User", Email: "db@example.com"}
	user.ID = 1
	mockRepo.On("FindByID", ctx, int64(1)).Return(user, nil)
	mockCache.On("Set", ctx, "user:1", mock.Anything, mock.Anything).Return(nil)

	result, err := svc.GetByID(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, "DB User", result.Name)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo, mockCache := newTestUserService(t)
	ctx := context.Background()

	mockCache.On("Get", ctx, "user:99", mock.Anything).Return(errors.New("miss"))
	mockRepo.On("FindByID", ctx, int64(99)).Return(nil, sql.ErrNoRows)

	result, err := svc.GetByID(ctx, 99)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

// ──────────────────────────────────────────────────────────────────────────────
// Create
// ──────────────────────────────────────────────────────────────────────────────

func TestUserService_Create_Success(t *testing.T) {
	svc, mockRepo, mockCache := newTestUserService(t)
	ctx := context.Background()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).
		Run(func(args mock.Arguments) {
			u := args.Get(1).(*model.User)
			u.ID = 1
		}).
		Return(nil)
	mockCache.On("DeleteByPrefix", ctx, userListCachePrefix).Return(nil)

	req := dto.CreateUserRequest{Name: "New", Email: "new@example.com"}
	result, err := svc.Create(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Create_Error(t *testing.T) {
	svc, mockRepo, _ := newTestUserService(t)
	ctx := context.Background()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(errors.New("dup"))

	req := dto.CreateUserRequest{Name: "Dup", Email: "dup@example.com"}
	result, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ──────────────────────────────────────────────────────────────────────────────
// Update
// ──────────────────────────────────────────────────────────────────────────────

func TestUserService_Update_Success(t *testing.T) {
	svc, mockRepo, mockCache := newTestUserService(t)
	ctx := context.Background()

	existing := &model.User{Name: "Old", Email: "old@example.com"}
	existing.ID = 1

	mockRepo.On("FindByID", ctx, int64(1)).Return(existing, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*model.User")).Return(nil)
	mockCache.On("Delete", ctx, "user:1").Return(nil)
	mockCache.On("DeleteByPrefix", ctx, userListCachePrefix).Return(nil)

	req := dto.UpdateUserRequest{Name: "New"}
	result, err := svc.Update(ctx, 1, req)

	require.NoError(t, err)
	assert.Equal(t, "New", result.Name)
	assert.Equal(t, "old@example.com", result.Email)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestUserService_Update_NotFound(t *testing.T) {
	svc, mockRepo, _ := newTestUserService(t)
	ctx := context.Background()

	mockRepo.On("FindByID", ctx, int64(99)).Return(nil, sql.ErrNoRows)

	req := dto.UpdateUserRequest{Name: "X"}
	result, err := svc.Update(ctx, 99, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ──────────────────────────────────────────────────────────────────────────────
// Delete
// ──────────────────────────────────────────────────────────────────────────────

func TestUserService_Delete_Success(t *testing.T) {
	svc, mockRepo, mockCache := newTestUserService(t)
	ctx := context.Background()

	mockRepo.On("Delete", ctx, int64(1)).Return(nil)
	mockCache.On("Delete", ctx, "user:1").Return(nil)
	mockCache.On("DeleteByPrefix", ctx, userListCachePrefix).Return(nil)

	err := svc.Delete(ctx, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestUserService_Delete_Error(t *testing.T) {
	svc, mockRepo, _ := newTestUserService(t)
	ctx := context.Background()

	mockRepo.On("Delete", ctx, int64(1)).Return(errors.New("db error"))

	err := svc.Delete(ctx, 1)

	assert.Error(t, err)
}
