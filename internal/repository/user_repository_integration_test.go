//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"service-app/internal/model"
	"service-app/test/integration/testhelper"
)

// Run with: go test -tags=integration ./internal/repository/...

func TestUserRepository_Integration_Create(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_user")
		testhelper.TeardownTestDB(t, db)
	})

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{Name: "John Doe", Email: "john@example.com"}
	err := repo.Create(ctx, user)

	require.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.Equal(t, "John Doe", user.Name)
}

func TestUserRepository_Integration_FindByID(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_user")
		testhelper.TeardownTestDB(t, db)
	})

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{Name: "Jane", Email: "jane@example.com"}
	require.NoError(t, repo.Create(ctx, user))

	found, err := repo.FindByID(ctx, user.ID)

	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "Jane", found.Name)
}

func TestUserRepository_Integration_FindByID_NotFound(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := NewUserRepository(db)
	ctx := context.Background()

	found, err := repo.FindByID(ctx, 99999)

	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_Integration_FindAll(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_user")
		testhelper.TeardownTestDB(t, db)
	})

	repo := NewUserRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &model.User{Name: "A", Email: "a@example.com"}))
	require.NoError(t, repo.Create(ctx, &model.User{Name: "B", Email: "b@example.com"}))

	users, err := repo.FindAll(ctx)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 2)
}

func TestUserRepository_Integration_Update(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_user")
		testhelper.TeardownTestDB(t, db)
	})

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{Name: "Original", Email: "orig@example.com"}
	require.NoError(t, repo.Create(ctx, user))

	user.Name = "Updated"
	require.NoError(t, repo.Update(ctx, user))

	updated, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
}

func TestUserRepository_Integration_Delete(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_user")
		testhelper.TeardownTestDB(t, db)
	})

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{Name: "ToDelete", Email: "del@example.com"}
	require.NoError(t, repo.Create(ctx, user))

	require.NoError(t, repo.Delete(ctx, user.ID))

	found, err := repo.FindByID(ctx, user.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
}
