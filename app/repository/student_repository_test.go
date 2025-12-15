package repository

import (
	models "achievement_backend/app/model"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupStudentRepoTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, StudentRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock db: %v", err)
	}

	repo := NewStudentRepository(db)
	return db, mock, repo
}


func TestStudentRepository_GetByID(t *testing.T) {
	db, mock, repo := setupStudentRepoTest(t)
	defer db.Close()

	advisorID := "advisor-uuid"
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"student_id",
		"program_study",
		"academic_year",
		"advisor_id",
		"full_name",
		"created_at",
	}).AddRow(
		"student-uuid",
		"user-uuid",
		"20231234",
		"Informatika",
		"2023",
		advisorID,
		"Cindy Permatasari Lubis",
		now,
	)

	mock.ExpectQuery("FROM students").
		WithArgs("student-uuid").
		WillReturnRows(rows)

	result, err := repo.GetByID("student-uuid")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "20231234", result.StudentID)
	assert.NotNil(t, result.AdvisorID)
	assert.Equal(t, advisorID, *result.AdvisorID)
}


func TestStudentRepository_Create(t *testing.T) {
	db, mock, repo := setupStudentRepoTest(t)
	defer db.Close()

	advisorID := "advisor-uuid"
	now := time.Now()

	req := models.CreateStudentRequest{
		UserID:       "user-uuid",
		StudentID:    "20231234",
		ProgramStudy: "Informatika",
		AcademicYear: "2023",
		AdvisorID:    &advisorID,
	}

	mock.ExpectQuery("INSERT INTO students").
		WithArgs(
			req.UserID,
			req.StudentID,
			req.ProgramStudy,
			req.AcademicYear,
			req.AdvisorID,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow("student-uuid"),
		)

	mock.ExpectQuery("FROM students").
		WithArgs("student-uuid").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"user_id",
			"student_id",
			"program_study",
			"academic_year",
			"advisor_id",
			"full_name",
			"created_at",
		}).AddRow(
			"student-uuid",
			"user-uuid",
			"20231234",
			"Informatika",
			"2023",
			advisorID,
			"Cindy Permatasari Lubis",
			now,
		))

	result, err := repo.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "20231234", result.StudentID)
	assert.NotNil(t, result.AdvisorID)
	assert.Equal(t, advisorID, *result.AdvisorID)
}


func TestStudentRepository_Update(t *testing.T) {
	db, mock, repo := setupStudentRepoTest(t)
	defer db.Close()

	now := time.Now()

	req := models.UpdateStudentRequest{
		UserID:       "user-uuid",
		StudentID:    "20239999",
		ProgramStudy: "Sistem Informasi",
		AcademicYear: "2024",
	}

	mock.ExpectExec("UPDATE students").
		WithArgs(
			req.UserID,
			req.StudentID,
			req.ProgramStudy,
			req.AcademicYear,
			"student-uuid",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("FROM students").
		WithArgs("student-uuid").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"user_id",
			"student_id",
			"program_study",
			"academic_year",
			"advisor_id",
			"full_name",
			"created_at",
		}).AddRow(
			"student-uuid",
			req.UserID,
			req.StudentID,
			req.ProgramStudy,
			req.AcademicYear,
			nil, // advisor_id NULL
			"Cindy Permatasari Lubis",
			now,
		))

	result, err := repo.Update("student-uuid", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.AdvisorID)
}


func TestStudentRepository_UpdateAdvisor(t *testing.T) {
	db, mock, repo := setupStudentRepoTest(t)
	defer db.Close()

	advisorID := "advisor-new"

	mock.ExpectExec("UPDATE students").
		WithArgs(advisorID, "student-uuid").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.UpdateAdvisor("student-uuid", advisorID)

	assert.NoError(t, err)
}
