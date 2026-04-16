package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"service-app/internal/cache"
	"service-app/internal/dto"
	"service-app/internal/helpers"
	"service-app/internal/model"
	"service-app/internal/repository"
	"service-app/internal/structs"
	"service-app/pkg/apperror"
)

const userListCachePrefix = "users:list:"

// UserService defines the business-logic interface for users.
type UserService interface {
	GetAll(ctx context.Context, params structs.ListParams) (*dto.PaginatedResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.UserResponse, error)
	Create(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error)
	Update(ctx context.Context, id int64, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	Delete(ctx context.Context, id int64) error
}

// userService implements UserService.
type userService struct {
	repo   repository.UserRepository
	cache  cache.Cache
	logger *slog.Logger
}

// NewUserService creates a new UserService with injected dependencies.
func NewUserService(repo repository.UserRepository, cache cache.Cache, logger *slog.Logger) UserService {
	return &userService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (s *userService) GetAll(ctx context.Context, params structs.ListParams) (*dto.PaginatedResponse, error) {
	// Build cache key from all query parameters
	cacheKey := buildUserListCacheKey(params)

	// Try cache first
	var cached dto.PaginatedResponse
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	users, total, err := s.repo.FindPaginated(ctx, params)
	if err != nil {
		s.logger.Error("failed to get paginated users", "error", err)
		return nil, apperror.NewInternal(err)
	}

	result := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, toUserResponse(u))
	}

	meta := helpers.BuildPaginationMeta(len(result), total, params.Pagination.Page, params.Pagination.Limit)
	resp := dto.PaginatedResponse{
		Data: result,
		Meta: meta,
	}

	// Cache with page-based TTL
	ttl := helpers.GetPaginationTTL(params.Pagination.Page)
	if cacheErr := s.cache.Set(ctx, cacheKey, resp, ttl); cacheErr != nil {
		s.logger.Warn("failed to set user list cache", "key", cacheKey, "error", cacheErr)
	}

	return &resp, nil
}

func (s *userService) GetByID(ctx context.Context, id int64) (*dto.UserResponse, error) {
	// Try cache first (cache-aside pattern)
	cacheKey := userCacheKey(id)
	var cached dto.UserResponse
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFound(fmt.Sprintf("user with id %d not found", id))
		}
		s.logger.Error("failed to get user by id", "id", id, "error", err)
		return nil, apperror.NewInternal(err)
	}

	resp := toUserResponse(*user)

	// Populate cache
	ttl := helpers.GetPaginationTTL(1) // same TTL as page 1
	if cacheErr := s.cache.Set(ctx, cacheKey, resp, ttl); cacheErr != nil {
		s.logger.Warn("failed to set user cache", "id", id, "error", cacheErr)
	}

	return &resp, nil
}

func (s *userService) Create(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	user := &model.User{
		Name:   req.Name,
		Email:  req.Email,
		RoleID: req.RoleID,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", "error", err)
		return nil, apperror.NewInternal(err)
	}

	// Invalidate list cache
	if cacheErr := s.cache.DeleteByPrefix(ctx, userListCachePrefix); cacheErr != nil {
		s.logger.Warn("failed to invalidate user list cache after create", "error", cacheErr)
	}

	resp := toUserResponse(*user)
	return &resp, nil
}

func (s *userService) Update(ctx context.Context, id int64, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFound(fmt.Sprintf("user with id %d not found", id))
		}
		s.logger.Error("failed to find user for update", "id", id, "error", err)
		return nil, apperror.NewInternal(err)
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.RoleID != nil {
		user.RoleID = req.RoleID
	}

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update user", "id", id, "error", err)
		return nil, apperror.NewInternal(err)
	}

	// Invalidate single entity cache
	if cacheErr := s.cache.Delete(ctx, userCacheKey(id)); cacheErr != nil {
		s.logger.Warn("failed to invalidate user cache", "id", id, "error", cacheErr)
	}
	// Invalidate list cache
	if cacheErr := s.cache.DeleteByPrefix(ctx, userListCachePrefix); cacheErr != nil {
		s.logger.Warn("failed to invalidate user list cache after update", "error", cacheErr)
	}

	resp := toUserResponse(*user)
	return &resp, nil
}

func (s *userService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete user", "id", id, "error", err)
		return apperror.NewInternal(err)
	}

	// Invalidate single entity cache
	if cacheErr := s.cache.Delete(ctx, userCacheKey(id)); cacheErr != nil {
		s.logger.Warn("failed to invalidate user cache after delete", "id", id, "error", cacheErr)
	}
	// Invalidate list cache
	if cacheErr := s.cache.DeleteByPrefix(ctx, userListCachePrefix); cacheErr != nil {
		s.logger.Warn("failed to invalidate user list cache after delete", "error", cacheErr)
	}

	return nil
}

// toUserResponse converts a model.User to a dto.UserResponse.
func toUserResponse(u model.User) dto.UserResponse {
	resp := dto.UserResponse{
		ID:     u.ID,
		Name:   u.Name,
		Email:  u.Email,
		RoleID: u.RoleID,
	}
	if u.Role != nil {
		resp.Role = &dto.RoleResponse{
			ID:       u.Role.ID,
			RoleName: u.Role.RoleName,
			RoleDesc: u.Role.RoleDesc,
			RoleCode: u.Role.RoleCode,
			Status:   u.Role.Status,
		}
	}
	return resp
}

// userCacheKey generates the Redis cache key for a single user.
func userCacheKey(id int64) string {
	return fmt.Sprintf("user:%d", id)
}

// buildUserListCacheKey generates a deterministic cache key from list parameters.
func buildUserListCacheKey(params structs.ListParams) string {
	// Sort orders for deterministic key
	orderParts := make([]string, 0, len(params.Orders))
	for _, o := range params.Orders {
		orderParts = append(orderParts, o.Column+"-"+o.Direction)
	}
	sort.Strings(orderParts)

	return fmt.Sprintf("%sp%d:l%d:s:%s:o:%s",
		userListCachePrefix,
		params.Pagination.Page,
		params.Pagination.Limit,
		params.Search,
		strings.Join(orderParts, ","),
	)
}
