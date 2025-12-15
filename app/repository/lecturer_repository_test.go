package repository

import (
	"regexp"
	"testing"
	"time"

	models "achievement_backend/app/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestLecturerRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	uuidUser := "123e4567-e89b-12d3-a456-426614174000"
	uuidLecturer := "223e4567-e89b-12d3-a456-426614174001"

	rows := sqlmock.NewRows([]string{"id", "user_id", "lecturer_id", "department", "created_at"}).
		AddRow(uuidLecturer, uuidUser, "DSN001", "Teknik", time.Now())

	mock.ExpectQuery("SELECT id, user_id, lecturer_id, department, created_at FROM lecturers ORDER BY created_at DESC").
		WillReturnRows(rows)

	repo := NewLecturerRepository(db)
	result, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, uuidUser, result[0].UserID)
	assert.Equal(t, uuidLecturer, result[0].ID)
	assert.Equal(t, "DSN001", result[0].LecturerID)
}

func TestLecturerRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	uuidUser := "123e4567-e89b-12d3-a456-426614174000"
	uuidLecturer := "223e4567-e89b-12d3-a456-426614174001"
	createdAt := time.Now()

	mock.ExpectQuery("SELECT id, user_id, lecturer_id, department, created_at FROM lecturers WHERE id = \\$1").
		WithArgs(uuidLecturer).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "lecturer_id", "department", "created_at"}).
			AddRow(uuidLecturer, uuidUser, "DSN001", "TI", createdAt))

	repo := NewLecturerRepository(db)
	l, err := repo.GetByID(uuidLecturer)
	assert.NoError(t, err)
	assert.Equal(t, "DSN001", l.LecturerID)
	assert.Equal(t, uuidUser, l.UserID)
}

func TestLecturerRepository_GetByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	uuidUser := "123e4567-e89b-12d3-a456-426614174000"
	uuidLecturer := "223e4567-e89b-12d3-a456-426614174001"
	createdAt := time.Now()

	mock.ExpectQuery("SELECT id, user_id, lecturer_id, department, created_at FROM lecturers WHERE user_id = \\$1").
		WithArgs(uuidUser).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "lecturer_id", "department", "created_at"}).
			AddRow(uuidLecturer, uuidUser, "DSN001", "TI", createdAt))

	repo := NewLecturerRepository(db)
	l, err := repo.GetByUserID(uuidUser)
	assert.NoError(t, err)
	assert.Equal(t, "TI", l.Department)
	assert.Equal(t, uuidUser, l.UserID)
	assert.Equal(t, uuidLecturer, l.ID)
}

func TestLecturerRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	uuidUser := "123e4567-e89b-12d3-a456-426614174000"
	uuidLecturer := "223e4567-e89b-12d3-a456-426614174001"

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO lecturers (user_id, lecturer_id, department) VALUES ($1, $2, $3) RETURNING id, created_at`)).
		WithArgs(uuidUser, "lect-1", "TI").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(uuidLecturer, time.Now()))

	mock.ExpectQuery("SELECT id, user_id, lecturer_id, department, created_at FROM lecturers WHERE id = \\$1").
		WithArgs(uuidLecturer).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "lecturer_id", "department", "created_at"}).
			AddRow(uuidLecturer, uuidUser, "lect-1", "TI", time.Now()))

	repo := NewLecturerRepository(db)
	l, err := repo.Create(models.CreateLecturerRequest{
		UserID:     uuidUser,
		LecturerID: "lect-1",
		Department: "TI",
	})
	assert.NoError(t, err)
	assert.Equal(t, "lect-1", l.LecturerID)
	assert.Equal(t, uuidUser, l.UserID)
	assert.Equal(t, uuidLecturer, l.ID)
}

func TestLecturerRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	uuidUser := "123e4567-e89b-12d3-a456-426614174000"
	uuidLecturer := "223e4567-e89b-12d3-a456-426614174001"

	mock.ExpectExec("UPDATE lecturers SET lecturer_id=\\$1, department=\\$2 WHERE id = \\$3").
		WithArgs("lect-2", "SI", uuidLecturer).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT id, user_id, lecturer_id, department, created_at FROM lecturers WHERE id = \\$1").
		WithArgs(uuidLecturer).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "lecturer_id", "department", "created_at"}).
			AddRow(uuidLecturer, uuidUser, "lect-2", "SI", time.Now()))

	repo := NewLecturerRepository(db)
	l, err := repo.Update(uuidLecturer, models.UpdateLecturerRequest{
		LecturerID: "lect-2",
		Department: "SI",
	})
	assert.NoError(t, err)
	assert.Equal(t, "lect-2", l.LecturerID)
	assert.Equal(t, uuidLecturer, l.ID)
}

func TestLecturerRepository_GetByLecturerID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	uuidUser := "123e4567-e89b-12d3-a456-426614174000"
	uuidLecturer := "223e4567-e89b-12d3-a456-426614174001"
	createdAt := time.Now()

	mock.ExpectQuery("SELECT id, user_id, lecturer_id, department, created_at FROM lecturers WHERE lecturer_id = \\$1").
		WithArgs("DSN001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "lecturer_id", "department", "created_at"}).
			AddRow(uuidLecturer, uuidUser, "DSN001", "TI", createdAt))

	repo := NewLecturerRepository(db)
	l, err := repo.GetByLecturerID("DSN001")
	assert.NoError(t, err)
	assert.Equal(t, "DSN001", l.LecturerID)
	assert.Equal(t, uuidLecturer, l.ID)
	assert.Equal(t, uuidUser, l.UserID)
}
