package repository

import (
	models "achievement_backend/app/model"
	"database/sql"
	"time"
)

type UserRepository interface {
	GetAll() ([]models.User, error)
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Create(req models.CreateUserRequest) (*models.User, error)
	Update(id string, req models.UpdateUserRequest) (*models.User, error)
	UpdateRole(id string, roleID string) error
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
	SELECT u.id, u.username, u.email, u.password_hash, u.full_name,
			u.role_id, COALESCE(r.name, '') AS role_name,
			u.is_active, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN roles r ON r.id = u.role_id
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.PasswordHash,
			&u.FullName, &u.RoleID, &u.RoleName, &u.IsActive,
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
		SELECT 
			u.id, u.username, u.email, u.password_hash, u.full_name,
			u.role_id,
			COALESCE(r.name, '') AS role_name,
			u.is_active, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`, id)

	var u models.User
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.FullName, &u.RoleID, &u.RoleName,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	row := r.db.QueryRow(`
		SELECT 
			u.id, u.username, u.email, u.password_hash, u.full_name,
			u.role_id,
			COALESCE(r.name, '') AS role_name,
			u.is_active, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.email=$1
	`, email)

	var u models.User
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.FullName, &u.RoleID, &u.RoleName,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	row := r.db.QueryRow(`
		SELECT 
			u.id, u.username, u.email, u.password_hash, u.full_name,
			u.role_id,
			COALESCE(r.name, '') AS role_name,
			u.is_active, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.username=$1
	`, username)

	var u models.User
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.FullName, &u.RoleID, &u.RoleName,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) Create(req models.CreateUserRequest) (*models.User, error) {
	var id string
	var isActive bool
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(`
		INSERT INTO users (username, email, password_hash, full_name)
		VALUES ($1,$2,$3,$4)
		RETURNING id, is_active, created_at, updated_at`,
		req.Username,
		req.Email,
		req.PasswordHash,
		req.FullName,
	).Scan(&id, &isActive, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:           id,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
		FullName:     req.FullName,
		RoleID:       nil,
		IsActive:     isActive,
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
		req.Username,
		req.Email,
		req.FullName,
		req.RoleID,
		req.IsActive,
		id,
	).Scan(&updatedAt)

	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        id,
		Username:  req.Username,
		Email:     req.Email,
		FullName:  req.FullName,
		RoleID:    &req.RoleID,
		IsActive:  req.IsActive,
		UpdatedAt: updatedAt,
	}, nil
}

func (r *userRepository) UpdateRole(id string, roleID string) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET role_id = $1
		WHERE id = $2
	`, roleID, id)

	return err
}

func (r *userRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}
