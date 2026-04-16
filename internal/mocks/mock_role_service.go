package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"service-app/internal/dto"
	"service-app/internal/structs"
)

// MockRoleService is a testify mock implementation of service.RoleService.
type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) GetAll(ctx context.Context, params structs.ListParams) (*dto.PaginatedResponse, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaginatedResponse), args.Error(1)
}

func (m *MockRoleService) GetByID(ctx context.Context, id int64) (*dto.RoleResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RoleResponse), args.Error(1)
}

func (m *MockRoleService) Create(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RoleResponse), args.Error(1)
}

func (m *MockRoleService) Update(ctx context.Context, id int64, req dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RoleResponse), args.Error(1)
}

func (m *MockRoleService) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
