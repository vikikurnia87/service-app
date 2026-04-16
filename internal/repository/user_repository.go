package repository

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"

	"service-app/internal/model"
	"service-app/internal/structs"
)

// UserRepository defines the data-access interface for users.
type UserRepository interface {
	FindAll(ctx context.Context) ([]model.User, error)
	FindPaginated(ctx context.Context, params structs.ListParams) ([]model.User, int, error)
	FindByID(ctx context.Context, id int64) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int64) error
}

// userRepository implements UserRepository using Bun ORM.
type userRepository struct {
	db *bun.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *bun.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindAll(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.NewSelect().
		Model(&users).
		OrderExpr("id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.FindAll: %w", err)
	}
	return users, nil
}

func (r *userRepository) FindPaginated(ctx context.Context, params structs.ListParams) ([]model.User, int, error) {
	var users []model.User
	q := r.db.NewSelect().Model(&users)

	// Search filter: ILIKE on name and email
	if params.Search != "" {
		search := "%" + params.Search + "%"
		q = q.Where("(tu.name ILIKE ? OR tu.email ILIKE ?)", search, search)
	}

	// Count total matching records (before limit/offset)
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("repository.FindPaginated count: %w", err)
	}

	// Apply ordering
	for _, o := range params.Orders {
		q = q.OrderExpr(fmt.Sprintf("%s %s", bun.Ident(o.Column), o.Direction))
	}

	// Apply pagination
	q = q.Limit(params.Pagination.Limit).Offset(params.Pagination.Offset)

	err = q.Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("repository.FindPaginated: %w", err)
	}

	return users, total, nil
}

func (r *userRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.FindByID: %w", err)
	}
	return user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.NewInsert().
		Model(user).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("repository.Create: %w", err)
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	_, err := r.db.NewUpdate().
		Model(user).
		WherePK().
		OmitZero().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("repository.Update: %w", err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*model.User)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("repository.Delete: %w", err)
	}
	return nil
}
