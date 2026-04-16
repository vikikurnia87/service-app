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

const roleListCachePrefix = "roles:list:"

// RoleService defines the business-logic interface for roles.
type RoleService interface {
	GetAll(ctx context.Context, params structs.ListParams) (*dto.PaginatedResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.RoleResponse, error)
	Create(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error)
	Update(ctx context.Context, id int64, req dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	Delete(ctx context.Context, id int64) error
}

// roleService implements RoleService.
type roleService struct {
	repo   repository.RoleRepository
	cache  cache.Cache
	logger *slog.Logger
}

// NewRoleService creates a new RoleService with injected dependencies.
func NewRoleService(repo repository.RoleRepository, cache cache.Cache, logger *slog.Logger) RoleService {
	return &roleService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (s *roleService) GetAll(ctx context.Context, params structs.ListParams) (*dto.PaginatedResponse, error) {
	// Build cache key from all query parameters
	cacheKey := buildRoleListCacheKey(params)

	// Try cache first
	var cached dto.PaginatedResponse
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	roles, total, err := s.repo.FindPaginated(ctx, params)
	if err != nil {
		s.logger.Error("failed to get paginated roles", "error", err)
		return nil, apperror.NewInternal(err)
	}

	result := make([]dto.RoleResponse, 0, len(roles))
	for _, r := range roles {
		result = append(result, toRoleResponse(r))
	}

	meta := helpers.BuildPaginationMeta(len(result), total, params.Pagination.Page, params.Pagination.Limit)
	resp := dto.PaginatedResponse{
		Data: result,
		Meta: meta,
	}

	// Cache with page-based TTL
	ttl := helpers.GetPaginationTTL(params.Pagination.Page)
	if cacheErr := s.cache.Set(ctx, cacheKey, resp, ttl); cacheErr != nil {
		s.logger.Warn("failed to set role list cache", "key", cacheKey, "error", cacheErr)
	}

	return &resp, nil
}

func (s *roleService) GetByID(ctx context.Context, id int64) (*dto.RoleResponse, error) {
	cacheKey := roleCacheKey(id)
	var cached dto.RoleResponse
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFound(fmt.Sprintf("role with id %d not found", id))
		}
		s.logger.Error("failed to get role by id", "id", id, "error", err)
		return nil, apperror.NewInternal(err)
	}

	resp := toRoleResponse(*role)

	ttl := helpers.GetPaginationTTL(1) // same TTL as page 1
	if cacheErr := s.cache.Set(ctx, cacheKey, resp, ttl); cacheErr != nil {
		s.logger.Warn("failed to set role cache", "id", id, "error", cacheErr)
	}

	return &resp, nil
}

func (s *roleService) Create(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	role := &model.Role{
		RoleName: req.RoleName,
		RoleDesc: req.RoleDesc,
		RoleCode: req.RoleCode,
	}

	if err := s.repo.Create(ctx, role); err != nil {
		s.logger.Error("failed to create role", "error", err)
		return nil, apperror.NewInternal(err)
	}

	// Invalidate list cache
	if cacheErr := s.cache.DeleteByPrefix(ctx, roleListCachePrefix); cacheErr != nil {
		s.logger.Warn("failed to invalidate role list cache after create", "error", cacheErr)
	}

	resp := toRoleResponse(*role)
	return &resp, nil
}

func (s *roleService) Update(ctx context.Context, id int64, req dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFound(fmt.Sprintf("role with id %d not found", id))
		}
		s.logger.Error("failed to find role for update", "id", id, "error", err)
		return nil, apperror.NewInternal(err)
	}

	if req.RoleName != "" {
		role.RoleName = req.RoleName
	}
	if req.RoleDesc != "" {
		role.RoleDesc = req.RoleDesc
	}
	if req.Status != nil {
		role.Status = *req.Status
	}

	if err := s.repo.Update(ctx, role); err != nil {
		s.logger.Error("failed to update role", "id", id, "error", err)
		return nil, apperror.NewInternal(err)
	}

	// Invalidate single entity cache
	if cacheErr := s.cache.Delete(ctx, roleCacheKey(id)); cacheErr != nil {
		s.logger.Warn("failed to invalidate role cache", "id", id, "error", cacheErr)
	}
	// Invalidate list cache
	if cacheErr := s.cache.DeleteByPrefix(ctx, roleListCachePrefix); cacheErr != nil {
		s.logger.Warn("failed to invalidate role list cache after update", "error", cacheErr)
	}

	resp := toRoleResponse(*role)
	return &resp, nil
}

func (s *roleService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete role", "id", id, "error", err)
		return apperror.NewInternal(err)
	}

	// Invalidate single entity cache
	if cacheErr := s.cache.Delete(ctx, roleCacheKey(id)); cacheErr != nil {
		s.logger.Warn("failed to invalidate role cache after delete", "id", id, "error", cacheErr)
	}
	// Invalidate list cache
	if cacheErr := s.cache.DeleteByPrefix(ctx, roleListCachePrefix); cacheErr != nil {
		s.logger.Warn("failed to invalidate role list cache after delete", "error", cacheErr)
	}

	return nil
}

// toRoleResponse converts a model.Role to a dto.RoleResponse.
func toRoleResponse(r model.Role) dto.RoleResponse {
	return dto.RoleResponse{
		ID:       r.ID,
		RoleName: r.RoleName,
		RoleDesc: r.RoleDesc,
		RoleCode: r.RoleCode,
		Status:   r.Status,
	}
}

// roleCacheKey generates the Redis cache key for a single role.
func roleCacheKey(id int64) string {
	return fmt.Sprintf("role:%d", id)
}

// buildRoleListCacheKey generates a deterministic cache key from list parameters.
func buildRoleListCacheKey(params structs.ListParams) string {
	// Sort orders for deterministic key
	orderParts := make([]string, 0, len(params.Orders))
	for _, o := range params.Orders {
		orderParts = append(orderParts, o.Column+"-"+o.Direction)
	}
	sort.Strings(orderParts)

	return fmt.Sprintf("%sp%d:l%d:s:%s:o:%s",
		roleListCachePrefix,
		params.Pagination.Page,
		params.Pagination.Limit,
		params.Search,
		strings.Join(orderParts, ","),
	)
}
