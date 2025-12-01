package repository

import (
	models "achievement_backend/app/model"
	"database/sql"
	"time"
)

type StudentRepository interface {
	GetAll() ([]models.Student, error)
	GetByID(id string) (*models.Student, error)
	GetByStudentID(studentID string) (*models.Student, error)
	GetByUserID(userID string) (*models.Student, error)
	GetByAdvisorID(advisorID string) ([]models.Student, error)
	Create(req models.CreateStudentRequest) (*models.Student, error)
	Update(id string, req models.UpdateStudentRequest) (*models.Student, error)
	Delete(id string) error
}

type studentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) StudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) GetAll() ([]models.Student, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.StudentID,
			&s.ProgramStudy,
			&s.AcademicYear,
			&s.AdvisorID,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}

		list = append(list, s)
	}

	return list, nil
}

func (r *studentRepository) GetByID(id string) (*models.Student, error) {
    var s models.Student

    row := r.db.QueryRow(`
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
        FROM students
        WHERE id = $1
    `, id)

    err := row.Scan(
        &s.ID,
        &s.UserID,
        &s.StudentID,
        &s.ProgramStudy,
        &s.AcademicYear,
        &s.AdvisorID,
        &s.CreatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, nil // FIX
    }
    if err != nil {
        return nil, err
    }

    return &s, nil
}

func (r *studentRepository) GetByStudentID(studentID string) (*models.Student, error) {
    var s models.Student

    row := r.db.QueryRow(`
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
        FROM students
        WHERE student_id = $1
    `, studentID)

    err := row.Scan(
        &s.ID,
        &s.UserID,
        &s.StudentID,
        &s.ProgramStudy,
        &s.AcademicYear,
        &s.AdvisorID,
        &s.CreatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, nil // FIX
    }
    if err != nil {
        return nil, err
    }

    return &s, nil
}

func (r *studentRepository) GetByUserID(userID string) (*models.Student, error) {
    var s models.Student

    row := r.db.QueryRow(`
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
        FROM students
        WHERE user_id = $1
    `, userID)

    err := row.Scan(
        &s.ID,
        &s.UserID,
        &s.StudentID,
        &s.ProgramStudy,
        &s.AcademicYear,
        &s.AdvisorID,
        &s.CreatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, nil // FIX
    }
    if err != nil {
        return nil, err
    }

    return &s, nil
}

func (r *studentRepository) GetByAdvisorID(advisorID string) ([]models.Student, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE advisor_id = $1
		ORDER BY created_at DESC
	`, advisorID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.StudentID,
			&s.ProgramStudy,
			&s.AcademicYear,
			&s.AdvisorID,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}

		list = append(list, s)
	}

	return list, nil
}

func (r *studentRepository) Create(req models.CreateStudentRequest) (*models.Student, error) {
	var id string

	err := r.db.QueryRow(`
		INSERT INTO students (user_id, student_id, program_study, academic_year, advisor_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, req.UserID, req.StudentID, req.ProgramStudy, req.AcademicYear, req.AdvisorID, time.Now()).Scan(&id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *studentRepository) Update(id string, req models.UpdateStudentRequest) (*models.Student, error) {
	result, err := r.db.Exec(`
		UPDATE students
		SET user_id=$1, student_id=$2, program_study=$3, academic_year=$4, advisor_id=$5
		WHERE id = $6
	`, req.UserID, req.StudentID, req.ProgramStudy, req.AcademicYear, req.AdvisorID, id)

	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, sql.ErrNoRows
	}

	return r.GetByID(id)
}

func (r *studentRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM students WHERE id=$1`, id)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
