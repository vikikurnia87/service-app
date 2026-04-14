package repository

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"

	"service-app/internal/model"
)

// RoleRepository defines the data-access interface for roles.
type RoleRepository interface {
	FindAll(ctx context.Context) ([]model.Role, error)
	FindByID(ctx context.Context, id int64) (*model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id int64) error
}

// roleRepository implements RoleRepository using Bun ORM.
type roleRepository struct {
	db *bun.DB
}

// NewRoleRepository creates a new RoleRepository.
func NewRoleRepository(db *bun.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) FindAll(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.NewSelect().
		Model(&roles).
		OrderExpr("id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.Role.FindAll: %w", err)
	}
	return roles, nil
}

func (r *roleRepository) FindByID(ctx context.Context, id int64) (*model.Role, error) {
	role := new(model.Role)
	err := r.db.NewSelect().
		Model(role).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.Role.FindByID: %w", err)
	}
	return role, nil
}

func (r *roleRepository) Create(ctx context.Context, role *model.Role) error {
	_, err := r.db.NewInsert().
		Model(role).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("repository.Role.Create: %w", err)
	}
	return nil
}

func (r *roleRepository) Update(ctx context.Context, role *model.Role) error {
	_, err := r.db.NewUpdate().
		Model(role).
		WherePK().
		OmitZero().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("repository.Role.Update: %w", err)
	}
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*model.Role)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("repository.Role.Delete: %w", err)
	}
	return nil
}
