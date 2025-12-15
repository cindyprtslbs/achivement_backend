package repository

import (
	"database/sql"
	"testing"
	"time"

	models "achievement_backend/app/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

//
// =======================================================
// SETUP
// =======================================================
//

func setupRoleRepoTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, RoleRepository) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	repo := NewRoleRepository(db)
	return db, mock, repo
}

//
// =======================================================
// GET ALL
// =======================================================
//

func TestRoleRepository_GetAll(t *testing.T) {
	db, mock, repo := setupRoleRepoTest(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "created_at",
	}).AddRow(
		"1", "Admin", "Administrator", time.Now(),
	)

	mock.ExpectQuery(`SELECT id, name, description, created_at FROM roles`).
		WillReturnRows(rows)

	roles, err := repo.GetAll()

	assert.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, "Admin", roles[0].Name)
}

//
// =======================================================
// GET BY ID
// =======================================================
//

func TestRoleRepository_GetByID(t *testing.T) {
	db, mock, repo := setupRoleRepoTest(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "created_at",
	}).AddRow(
		"1", "Student", "Mahasiswa", time.Now(),
	)

	mock.ExpectQuery(`FROM roles WHERE id = \$1`).
		WithArgs("1").
		WillReturnRows(rows)

	role, err := repo.GetByID("1")

	assert.NoError(t, err)
	assert.Equal(t, "Student", role.Name)
}

//
// =======================================================
// CREATE
// =======================================================
//

func TestRoleRepository_Create(t *testing.T) {
	db, mock, repo := setupRoleRepoTest(t)
	defer db.Close()

	mock.ExpectQuery(`INSERT INTO roles`).
		WithArgs("Lecturer", "Dosen", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("10"))

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "created_at",
	}).AddRow(
		"10", "Lecturer", "Dosen", time.Now(),
	)

	mock.ExpectQuery(`FROM roles WHERE id = \$1`).
		WithArgs("10").
		WillReturnRows(rows)

	role, err := repo.Create(models.CreateRoleRequest{
		Name:        "Lecturer",
		Description: "Dosen",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Lecturer", role.Name)
}

//
// =======================================================
// UPDATE
// =======================================================
//

func TestRoleRepository_Update(t *testing.T) {
	db, mock, repo := setupRoleRepoTest(t)
	defer db.Close()

	mock.ExpectExec(`UPDATE roles`).
		WithArgs("Admin Updated", "Updated Desc", "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "created_at",
	}).AddRow(
		"1", "Admin Updated", "Updated Desc", time.Now(),
	)

	mock.ExpectQuery(`FROM roles WHERE id = \$1`).
		WithArgs("1").
		WillReturnRows(rows)

	role, err := repo.Update("1", models.UpdateRoleRequest{
		Name:        "Admin Updated",
		Description: "Updated Desc",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Admin Updated", role.Name)
}

//
// =======================================================
// DELETE
// =======================================================
//

func TestRoleRepository_Delete(t *testing.T) {
	db, mock, repo := setupRoleRepoTest(t)
	defer db.Close()

	mock.ExpectExec(`DELETE FROM roles WHERE id=\$1`).
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete("1")

	assert.NoError(t, err)
}
