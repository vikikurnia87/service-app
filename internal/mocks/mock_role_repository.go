package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"service-app/internal/model"
	"service-app/internal/structs"
)

// MockRoleRepository is a testify mock implementation of repository.RoleRepository.
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) FindAll(ctx context.Context) ([]model.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRoleRepository) FindPaginated(ctx context.Context, params structs.ListParams) ([]model.Role, int, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]model.Role), args.Int(1), args.Error(2)
}

func (m *MockRoleRepository) FindByID(ctx context.Context, id int64) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleRepository) Create(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
