package repository

import (
	"database/sql"
	models "achievement_backend/app/model"
)

type AuthRepository interface {
	GetForLogin(identifier string) (*models.User, error)
}

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetForLogin(identifier string) (*models.User, error) {
	row := r.db.QueryRow(`
		SELECT 
			id, username, email, password_hash, full_name, 
			role_id, is_active, created_at, updated_at
		FROM users
		WHERE username = $1 OR email = $1
		LIMIT 1
	`, identifier)

	var u models.User

	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.RoleID,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
