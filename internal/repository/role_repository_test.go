package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"service-app/internal/model"
)

// ──────────────────────────────────────────────────────────────────────────────
// FindAll
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleRepository_FindAll(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRoleRepository(db)

	rows := sqlmock.NewRows([]string{"id", "role_name", "role_desc", "role_code", "status"}).
		AddRow(int64(1), "Admin", "Administrator", "ADMIN", int16(1)).
		AddRow(int64(2), "User", "Regular user", "USER", int16(1))

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	roles, err := repo.FindAll(context.Background())

	require.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.Equal(t, "Admin", roles[0].RoleName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_FindAll_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(assert.AnError)

	roles, err := repo.FindAll(context.Background())

	assert.Error(t, err)
	assert.Nil(t, roles)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ──────────────────────────────────────────────────────────────────────────────
// FindByID
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleRepository_FindByID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRoleRepository(db)

	rows := sqlmock.NewRows([]string{"id", "role_name", "role_desc", "role_code", "status"}).
		AddRow(int64(1), "Admin", "Administrator", "ADMIN", int16(1))

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	role, err := repo.FindByID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, "Admin", role.RoleName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_FindByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRoleRepository(db)

	rows := sqlmock.NewRows([]string{"id", "role_name", "role_desc", "role_code", "status"})
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	role, err := repo.FindByID(context.Background(), 99)

	assert.Error(t, err)
	assert.Nil(t, role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ──────────────────────────────────────────────────────────────────────────────
// Create
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("INSERT").WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow(int64(1)),
	)

	role := &model.Role{RoleName: "Admin", RoleDesc: "Administrator", RoleCode: "ADMIN"}
	err := repo.Create(context.Background(), role)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_Create_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("INSERT").WillReturnError(assert.AnError)

	role := &model.Role{RoleName: "Dup", RoleDesc: "Dup", RoleCode: "DUP"}
	err := repo.Create(context.Background(), role)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ──────────────────────────────────────────────────────────────────────────────
// Update
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleRepository_Update(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(0, 1))

	role := &model.Role{RoleName: "Updated", RoleDesc: "Updated desc", RoleCode: "UPD"}
	role.ID = 1
	err := repo.Update(context.Background(), role)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ──────────────────────────────────────────────────────────────────────────────
// Delete
// ──────────────────────────────────────────────────────────────────────────────

func TestRoleRepository_Delete(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
