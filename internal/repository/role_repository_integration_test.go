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

func TestRoleRepository_Integration_Create(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_role")
		testhelper.TeardownTestDB(t, db)
	})

	repo := NewRoleRepository(db)
	ctx := context.Background()

	role := &model.Role{RoleName: "Admin", RoleDesc: "Administrator", RoleCode: "ADMIN"}
	err := repo.Create(ctx, role)

	require.NoError(t, err)
	assert.NotZero(t, role.ID)
	assert.Equal(t, "Admin", role.RoleName)
}

func TestRoleRepository_Integration_FindByID(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_role")
		testhelper.TeardownTestDB(t, db)
	})

	repo := NewRoleRepository(db)
	ctx := context.Background()

	role := &model.Role{RoleName: "Editor", RoleDesc: "Content editor", RoleCode: "EDITOR"}
	require.NoError(t, repo.Create(ctx, role))

	found, err := repo.FindByID(ctx, role.ID)

	require.NoError(t, err)
	assert.Equal(t, role.ID, found.ID)
	assert.Equal(t, "Editor", found.RoleName)
	assert.Equal(t, "EDITOR", found.RoleCode)
}

func TestRoleRepository_Integration_FindAll(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_role")
		testhelper.TeardownTestDB(t, db)
	})

	repo := NewRoleRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &model.Role{RoleName: "A", RoleDesc: "a", RoleCode: "A"}))
	require.NoError(t, repo.Create(ctx, &model.Role{RoleName: "B", RoleDesc: "b", RoleCode: "B"}))

	roles, err := repo.FindAll(ctx)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(roles), 2)
}

func TestRoleRepository_Integration_Update(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_role")
		testhelper.TeardownTestDB(t, db)
	})

	repo := NewRoleRepository(db)
	ctx := context.Background()

	role := &model.Role{RoleName: "Original", RoleDesc: "orig", RoleCode: "ORIG"}
	require.NoError(t, repo.Create(ctx, role))

	role.RoleName = "Updated"
	require.NoError(t, repo.Update(ctx, role))

	updated, err := repo.FindByID(ctx, role.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.RoleName)
}

func TestRoleRepository_Integration_Delete(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() {
		testhelper.CleanTable(t, db, "t_role")
		testhelper.TeardownTestDB(t, db)
	})

	repo := NewRoleRepository(db)
	ctx := context.Background()

	role := &model.Role{RoleName: "ToDelete", RoleDesc: "del", RoleCode: "DEL"}
	require.NoError(t, repo.Create(ctx, role))

	require.NoError(t, repo.Delete(ctx, role.ID))

	found, err := repo.FindByID(ctx, role.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
}
