package repository

import (
	"database/sql"
	"testing"
	"time"

	models "achievement_backend/app/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, UserRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock db: %v", err)
	}

	repo := NewUserRepository(db)
	return db, mock, repo
}

// ==================== GET BY ID ====================

func TestUserRepository_GetByID_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash",
		"full_name", "role_id", "role_name",
		"is_active", "created_at", "updated_at",
	}).AddRow(
		"uuid-cindy-1", "cindy", "cindy@mail.com", "hashed",
		"Cindy Lubis", nil, "Mahasiswa",
		true, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT (.+) FROM users`).
		WithArgs("uuid-cindy-1").
		WillReturnRows(rows)

	user, err := repo.GetByID("uuid-cindy-1")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "cindy", user.Username)
	assert.Equal(t, "Mahasiswa", user.RoleName)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(`SELECT (.+) FROM users`).
		WithArgs("invalid-id").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByID("invalid-id")

	assert.Error(t, err)
	assert.Nil(t, user)
}

// ==================== CREATE ====================

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	req := models.CreateUserRequest{
		Username:     "cindy",
		Email:        "cindy@mail.com",
		PasswordHash: "hashed",
		FullName:     "Cindy Lubis",
	}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(req.Username, req.Email, req.PasswordHash, req.FullName).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "is_active", "created_at", "updated_at"}).
				AddRow("uuid-cindy-2", true, time.Now(), time.Now()),
		)

	user, err := repo.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "cindy", user.Username)
	assert.True(t, user.IsActive)
}

// ==================== UPDATE PARTIAL ====================

func TestUserRepository_UpdatePartial_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	user := &models.User{
		ID:       "uuid-cindy-3",
		Username: "cindy_updated",
		Email:    "cindy.updated@mail.com",
		FullName: "Cindy Lubis Updated",
		IsActive: true,
		RoleID:   nil,
	}

	mock.ExpectExec(`UPDATE users SET`).
		WithArgs(
			user.Username,
			user.Email,
			user.FullName,
			user.RoleID,
			user.IsActive,
			user.ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// GetByID dipanggil ulang
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash",
		"full_name", "role_id", "role_name",
		"is_active", "created_at", "updated_at",
	}).AddRow(
		user.ID, user.Username, user.Email, "hashed",
		user.FullName, nil, "Admin",
		true, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT (.+) FROM users`).
		WithArgs(user.ID).
		WillReturnRows(rows)

	result, err := repo.UpdatePartial(user)

	assert.NoError(t, err)
	assert.Equal(t, "cindy_updated", result.Username)
	assert.Equal(t, "Cindy Lubis Updated", result.FullName)
}

// ==================== UPDATE PASSWORD ====================

func TestUserRepository_UpdatePassword_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(`UPDATE users SET password_hash`).
		WithArgs("newhash", "uuid-cindy-4").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.UpdatePassword("uuid-cindy-4", "newhash")

	assert.NoError(t, err)
}

// ==================== DELETE ====================

func TestUserRepository_Delete_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(`DELETE FROM users`).
		WithArgs("uuid-cindy-5").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Delete("uuid-cindy-5")

	assert.NoError(t, err)
}
