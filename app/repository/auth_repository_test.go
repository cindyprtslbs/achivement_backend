package repository

import (
	"database/sql"
	"testing"
	"time"

	// models "achievement_backend/app/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupAuthRepo(t *testing.T) (*sql.DB, sqlmock.Sqlmock, AuthRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	repo := NewAuthRepository(db)
	return db, mock, repo
}

// =======================================================
// TEST SUCCESS LOGIN (USERNAME / EMAIL)
// =======================================================

func TestAuthRepository_GetForLogin_Success(t *testing.T) {
	db, mock, repo := setupAuthRepo(t)
	defer db.Close()

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"username",
		"email",
		"password_hash",
		"full_name",
		"role_id",
		"is_active",
		"created_at",
		"updated_at",
	}).AddRow(
		"1",
		"cindy",
		"cindy@mail.com",
		"$2a$10$hash",
		"cindy Doe",
		"role-1",
		true,
		now,
		now,
	)

	mock.ExpectQuery(`
		SELECT 
			id, username, email, password_hash, full_name, 
			role_id, is_active, created_at, updated_at
		FROM users
		WHERE username = \$1 OR email = \$1
		LIMIT 1
	`).WithArgs("cindy").WillReturnRows(rows)

	user, err := repo.GetForLogin("cindy")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "cindy", user.Username)
	assert.Equal(t, "cindy@mail.com", user.Email)
	assert.True(t, user.IsActive)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// =======================================================
// TEST USER NOT FOUND
// =======================================================

func TestAuthRepository_GetForLogin_NotFound(t *testing.T) {
	db, mock, repo := setupAuthRepo(t)
	defer db.Close()

	mock.ExpectQuery(`
		SELECT 
			id, username, email, password_hash, full_name, 
			role_id, is_active, created_at, updated_at
		FROM users
		WHERE username = \$1 OR email = \$1
		LIMIT 1
	`).WithArgs("unknown").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetForLogin("unknown")

	assert.Error(t, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}
