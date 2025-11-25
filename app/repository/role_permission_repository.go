package repository

import (
	"database/sql"
	models "achievement_backend/app/model"
)

type RolePermissionRepository interface {
	AssignPermission(roleID string, permissionID string) error
	RemovePermission(roleID string, permissionID string) error
	GetPermissionsByRole(roleID string) ([]models.Permission, error)
	GetRolesByPermission(permissionID string) ([]models.Role, error)
}

type rolePermissionRepository struct {
	db *sql.DB
}

func NewRolePermissionRepository(db *sql.DB) RolePermissionRepository {
	return &rolePermissionRepository{db: db}
}

func (r *rolePermissionRepository) AssignPermission(roleID string, permissionID string) error {
	_, err := r.db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, roleID, permissionID)

	return err
}

func (r *rolePermissionRepository) RemovePermission(roleID string, permissionID string) error {
	result, err := r.db.Exec(`
		DELETE FROM role_permissions
		WHERE role_id = $1 AND permission_id = $2
	`, roleID, permissionID)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *rolePermissionRepository) GetPermissionsByRole(roleID string) ([]models.Permission, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.name, p.resource, p.action, p.description
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.name ASC
	`, roleID)

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

func (r *rolePermissionRepository) GetRolesByPermission(permissionID string) ([]models.Role, error) {
	rows, err := r.db.Query(`
		SELECT r.id, r.name, r.description, r.created_at
		FROM roles r
		INNER JOIN role_permissions rp ON rp.role_id = r.id
		WHERE rp.permission_id = $1
		ORDER BY r.name ASC
	`, permissionID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Role
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

		list = append(list, role)
	}

	return list, nil
}
