//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"service-app/internal/model"
	"service-app/internal/repository"
	"service-app/test/integration/testhelper"
)

func TestUserRepository_Integration_Create(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{Name: "John Doe", Email: "john@example.com"}
	err := repo.Create(ctx, user)

	require.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.Equal(t, "John Doe", user.Name)

	// cleanup: hapus row yang dibuat test
	t.Cleanup(func() { testhelper.DeleteByID(t, db, "t_user", user.ID) })
}

func TestUserRepository_Integration_FindByID(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{Name: "Jane", Email: "jane@example.com"}
	require.NoError(t, repo.Create(ctx, user))
	t.Cleanup(func() { testhelper.DeleteByID(t, db, "t_user", user.ID) })

	found, err := repo.FindByID(ctx, user.ID)

	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "Jane", found.Name)
}

func TestUserRepository_Integration_FindByID_NotFound(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	found, err := repo.FindByID(ctx, 99999)

	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_Integration_FindAll(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	u1 := &model.User{Name: "A", Email: "a@example.com"}
	u2 := &model.User{Name: "B", Email: "b@example.com"}
	require.NoError(t, repo.Create(ctx, u1))
	require.NoError(t, repo.Create(ctx, u2))
	t.Cleanup(func() { testhelper.DeleteByIDs(t, db, "t_user", u1.ID, u2.ID) })

	users, err := repo.FindAll(ctx)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 2)
}

func TestUserRepository_Integration_Update(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{Name: "Original", Email: "orig@example.com"}
	require.NoError(t, repo.Create(ctx, user))
	t.Cleanup(func() { testhelper.DeleteByID(t, db, "t_user", user.ID) })

	user.Name = "Updated"
	require.NoError(t, repo.Update(ctx, user))

	updated, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
}

func TestUserRepository_Integration_Delete(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{Name: "ToDelete", Email: "del@example.com"}
	require.NoError(t, repo.Create(ctx, user))

	require.NoError(t, repo.Delete(ctx, user.ID))

	found, err := repo.FindByID(ctx, user.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
}
