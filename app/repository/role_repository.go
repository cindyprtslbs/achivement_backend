package repository

import (
	"database/sql"
	models "achievement_backend/app/model"
	"time"
)

type RoleRepository interface {
	GetAll() ([]models.Role, error)
	GetByID(id string) (*models.Role, error)
	Create(req models.CreateRoleRequest) (*models.Role, error)
	Update(id string, req models.UpdateRoleRequest) (*models.Role, error)
	Delete(id string) error
}

type roleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetAll() ([]models.Role, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, created_at 
		FROM roles
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.CreatedAt,
		); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *roleRepository) GetByID(id string) (*models.Role, error) {
	var role models.Role

	row := r.db.QueryRow(`
		SELECT id, name, description, created_at
		FROM roles
		WHERE id = $1
	`, id)

	err := row.Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *roleRepository) Create(req models.CreateRoleRequest) (*models.Role, error) {
	var id string

	err := r.db.QueryRow(`
		INSERT INTO roles (name, description, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`, req.Name, req.Description, time.Now()).Scan(&id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *roleRepository) Update(id string, req models.UpdateRoleRequest) (*models.Role, error) {
	result, err := r.db.Exec(`
		UPDATE roles 
		SET name=$1, description=$2
		WHERE id = $3
	`, req.Name, req.Description, id)

	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, sql.ErrNoRows
	}

	return r.GetByID(id)
}

func (r *roleRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM roles WHERE id=$1`, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
