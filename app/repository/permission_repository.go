package repository

import (
	"database/sql"
	models "achievement_backend/app/model"
)

type PermissionRepository interface {
	GetAll() ([]models.Permission, error)
	GetByID(id string) (*models.Permission, error)
	Create(req models.CreatePermissionRequest) (*models.Permission, error)
	Update(id string, req models.UpdatePermissionRequest) (*models.Permission, error)
	Delete(id string) error
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

func (r *permissionRepository) Create(req models.CreatePermissionRequest) (*models.Permission, error) {
	var id string

	err := r.db.QueryRow(`
		INSERT INTO permissions (name, resource, action, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, req.Name, req.Resource, req.Action, req.Description).Scan(&id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *permissionRepository) Update(id string, req models.UpdatePermissionRequest) (*models.Permission, error) {
	result, err := r.db.Exec(`
		UPDATE permissions
		SET name=$1, resource=$2, action=$3, description=$4
		WHERE id = $5
	`, req.Name, req.Resource, req.Action, req.Description, id)

	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, sql.ErrNoRows
	}

	return r.GetByID(id)
}

func (r *permissionRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM permissions WHERE id=$1`, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
