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

func TestRoleRepository_Integration_Create(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := repository.NewRoleRepository(db)
	ctx := context.Background()

	role := &model.Role{RoleName: "Admin", RoleDesc: "Administrator", RoleCode: "ADMIN"}
	err := repo.Create(ctx, role)

	require.NoError(t, err)
	assert.NotZero(t, role.ID)
	assert.Equal(t, "Admin", role.RoleName)

	t.Cleanup(func() { testhelper.DeleteByID(t, db, "t_role", role.ID) })
}

func TestRoleRepository_Integration_FindByID(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := repository.NewRoleRepository(db)
	ctx := context.Background()

	role := &model.Role{RoleName: "Editor", RoleDesc: "Content editor", RoleCode: "EDITOR"}
	require.NoError(t, repo.Create(ctx, role))
	t.Cleanup(func() { testhelper.DeleteByID(t, db, "t_role", role.ID) })

	found, err := repo.FindByID(ctx, role.ID)

	require.NoError(t, err)
	assert.Equal(t, role.ID, found.ID)
	assert.Equal(t, "Editor", found.RoleName)
	assert.Equal(t, "EDITOR", found.RoleCode)
}

func TestRoleRepository_Integration_FindAll(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := repository.NewRoleRepository(db)
	ctx := context.Background()

	r1 := &model.Role{RoleName: "A", RoleDesc: "a", RoleCode: "A"}
	r2 := &model.Role{RoleName: "B", RoleDesc: "b", RoleCode: "B"}
	require.NoError(t, repo.Create(ctx, r1))
	require.NoError(t, repo.Create(ctx, r2))
	t.Cleanup(func() { testhelper.DeleteByIDs(t, db, "t_role", r1.ID, r2.ID) })

	roles, err := repo.FindAll(ctx)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(roles), 2)
}

func TestRoleRepository_Integration_Update(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := repository.NewRoleRepository(db)
	ctx := context.Background()

	role := &model.Role{RoleName: "Original", RoleDesc: "orig", RoleCode: "ORIG"}
	require.NoError(t, repo.Create(ctx, role))
	t.Cleanup(func() { testhelper.DeleteByID(t, db, "t_role", role.ID) })

	role.RoleName = "Updated"
	require.NoError(t, repo.Update(ctx, role))

	updated, err := repo.FindByID(ctx, role.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.RoleName)
}

func TestRoleRepository_Integration_Delete(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	t.Cleanup(func() { testhelper.TeardownTestDB(t, db) })

	repo := repository.NewRoleRepository(db)
	ctx := context.Background()

	role := &model.Role{RoleName: "ToDelete", RoleDesc: "del", RoleCode: "DEL"}
	require.NoError(t, repo.Create(ctx, role))

	require.NoError(t, repo.Delete(ctx, role.ID))

	found, err := repo.FindByID(ctx, role.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
}
