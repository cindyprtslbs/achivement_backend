package repository

import (
	"database/sql"
	models "achievement_backend/app/model"
)

type PermissionRepository interface {
	GetAll() ([]models.Permission, error)
	GetByID(id string) (*models.Permission, error)
}

type permissionRepository struct {
	db *sql.DB
}

func NewPermissionRepository(db *sql.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) GetAll() ([]models.Permission, error) {
	rows, err := r.db.Query(`
		SELECT id, name, resource, action, description
		FROM permissions
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Permission
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Resource,
			&p.Action,
			&p.Description,
		); err != nil {
			return nil, err
		}
		list = append(list, p)
	}

	return list, nil
}

func (r *permissionRepository) GetByID(id string) (*models.Permission, error) {
	var p models.Permission

	row := r.db.QueryRow(`
		SELECT id, name, resource, action, description
		FROM permissions
		WHERE id = $1
	`, id)

	err := row.Scan(
		&p.ID,
		&p.Name,
		&p.Resource,
		&p.Action,
		&p.Description,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
