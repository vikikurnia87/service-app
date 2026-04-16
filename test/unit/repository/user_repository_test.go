package repository_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"service-app/internal/model"
	"service-app/internal/repository"
)

// ──────────────────────────────────────────────────────────────────────────────
// FindAll
// ──────────────────────────────────────────────────────────────────────────────

func TestUserRepository_FindAll(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "email"}).
		AddRow(int64(1), "Alice", "alice@example.com").
		AddRow(int64(2), "Bob", "bob@example.com")

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	users, err := repo.FindAll(context.Background())

	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "Alice", users[0].Name)
	assert.Equal(t, "Bob", users[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindAll_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(assert.AnError)

	users, err := repo.FindAll(context.Background())

	assert.Error(t, err)
	assert.Nil(t, users)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ──────────────────────────────────────────────────────────────────────────────
// FindByID
// ──────────────────────────────────────────────────────────────────────────────

func TestUserRepository_FindByID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "email"}).
		AddRow(int64(1), "Alice", "alice@example.com")

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	user, err := repo.FindByID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "Alice", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "email"})
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	user, err := repo.FindByID(context.Background(), 99)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ──────────────────────────────────────────────────────────────────────────────
// Create
// ──────────────────────────────────────────────────────────────────────────────

func TestUserRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewUserRepository(db)

	mock.ExpectQuery("INSERT").WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow(int64(1)),
	)

	user := &model.User{Name: "New User", Email: "new@example.com"}
	err := repo.Create(context.Background(), user)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewUserRepository(db)

	mock.ExpectQuery("INSERT").WillReturnError(assert.AnError)

	user := &model.User{Name: "Dup", Email: "dup@example.com"}
	err := repo.Create(context.Background(), user)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ──────────────────────────────────────────────────────────────────────────────
// Update
// ──────────────────────────────────────────────────────────────────────────────

func TestUserRepository_Update(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewUserRepository(db)

	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(0, 1))

	user := &model.User{Name: "Updated", Email: "updated@example.com"}
	user.ID = 1
	err := repo.Update(context.Background(), user)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ──────────────────────────────────────────────────────────────────────────────
// Delete
// ──────────────────────────────────────────────────────────────────────────────

func TestUserRepository_Delete(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewUserRepository(db)

	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewUserRepository(db)

	mock.ExpectExec("DELETE").WillReturnError(assert.AnError)

	err := repo.Delete(context.Background(), 1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
