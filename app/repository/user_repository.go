package repository

import (
	"database/sql"
	models "achievement_backend/app/model"
	"time"
)

type UserRepository interface {
	GetAll() ([]models.User, error)
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Create(req models.CreateUserRequest) (*models.User, error)
	Update(id string, req models.UpdateUserRequest) (*models.User, error)
	Delete(id string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAll() ([]models.User, error) {
	rows, err := r.db.Query(`
		SELECT id, username, email, password_hash, full_name, role_id, is_active,
			   created_at, updated_at
		FROM users 
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.PasswordHash,
			&u.FullName, &u.RoleID, &u.IsActive,
			&u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, u)
	}

	return list, nil
}

func (r *userRepository) GetByID(id string) (*models.User, error) {
	row := r.db.QueryRow(`
		SELECT id, username, email, password_hash, full_name, role_id, is_active,
			   created_at, updated_at
		FROM users WHERE id=$1`, id)

	var u models.User
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.FullName, &u.RoleID, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	row := r.db.QueryRow(`
		SELECT id, username, email, password_hash, full_name, role_id, is_active,
			   created_at, updated_at
		FROM users WHERE email=$1`, email)

	var u models.User
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.FullName, &u.RoleID, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	row := r.db.QueryRow(`
		SELECT id, username, email, password_hash, full_name, role_id, is_active,
			   created_at, updated_at
		FROM users WHERE username=$1`, username)

	var u models.User
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.FullName, &u.RoleID, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) Create(req models.CreateUserRequest) (*models.User, error) {
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(`
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active,
						   created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7, NOW(), NOW())
		RETURNING created_at, updated_at`,
		req.ID, req.Username, req.Email, req.PasswordHash,
		req.FullName, req.RoleID, true,
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:           req.ID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		IsActive:     true,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

func (r *userRepository) Update(id string, req models.UpdateUserRequest) (*models.User, error) {
	var updatedAt time.Time

	err := r.db.QueryRow(`
		UPDATE users SET
			username=$1,
			email=$2,
			full_name=$3,
			role_id=$4,
			is_active=$5,
			updated_at=NOW()
		WHERE id=$6
		RETURNING updated_at`,
		req.Username, req.Email, req.FullName,
		req.RoleID, id,
	).Scan(&updatedAt)

	if err != nil {
		return nil, err
	}

	updated := &models.User{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
		RoleID:   req.RoleID,
		UpdatedAt: updatedAt,
	}

	return updated, nil
}

func (r *userRepository) Delete(id string) error {
	_, err := r.db.Exec(`
	DELETE FROM users WHERE id=$1`, id)
	return err
}
