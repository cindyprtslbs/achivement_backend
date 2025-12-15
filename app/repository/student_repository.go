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
	UpdateAdvisor(id string, advisorID string) error
}

type studentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) StudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) GetAll() ([]models.Student, error) {
	rows, err := r.db.Query(`
		SELECT 
			s.id,
			s.user_id,
			s.student_id,
			s.program_study,
			s.academic_year,
			s.advisor_id,
			u.full_name,
			s.created_at
		FROM students s
		LEFT JOIN users u ON u.id = s.user_id
		ORDER BY s.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Student
	for rows.Next() {
		var s models.Student
		err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.StudentID,
			&s.ProgramStudy,
			&s.AcademicYear,
			&s.AdvisorID,
			&s.FullName,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *studentRepository) GetByID(id string) (*models.Student, error) {
	var s models.Student

	row := r.db.QueryRow(`
		SELECT
			s.id,
			s.user_id,
			s.student_id,
			s.program_study,
			s.academic_year,
			s.advisor_id,
			u.full_name,
			s.created_at
		FROM students s
		LEFT JOIN users u ON u.id = s.user_id
		WHERE s.id = $1
	`, id)

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.StudentID,
		&s.ProgramStudy,
		&s.AcademicYear,
		&s.AdvisorID,
		&s.FullName,
		&s.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *studentRepository) GetByStudentID(studentID string) (*models.Student, error) {
	var s models.Student

	row := r.db.QueryRow(`
		SELECT
			s.id,
			s.user_id,
			s.student_id,
			s.program_study,
			s.academic_year,
			s.advisor_id,
			u.full_name,
			s.created_at
		FROM students s
		LEFT JOIN users u ON u.id = s.user_id
		WHERE s.student_id = $1
	`, studentID)

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.StudentID,
		&s.ProgramStudy,
		&s.AcademicYear,
		&s.AdvisorID,
		&s.FullName,
		&s.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *studentRepository) GetByUserID(userID string) (*models.Student, error) {
	var s models.Student

	row := r.db.QueryRow(`
		SELECT
			s.id,
			s.user_id,
			s.student_id,
			s.program_study,
			s.academic_year,
			s.advisor_id,
			u.full_name,
			s.created_at
		FROM students s
		LEFT JOIN users u ON u.id = s.user_id
		WHERE s.user_id = $1
	`, userID)

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.StudentID,
		&s.ProgramStudy,
		&s.AcademicYear,
		&s.AdvisorID,
		&s.FullName,
		&s.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *studentRepository) GetByAdvisorID(advisorID string) ([]models.Student, error) {
	rows, err := r.db.Query(`
		SELECT 
			s.id,
			s.user_id,
			s.student_id,
			s.program_study,
			s.academic_year,
			s.advisor_id,
			u.full_name,
			s.created_at
		FROM students s
		LEFT JOIN users u ON u.id = s.user_id
		WHERE s.advisor_id = $1
		ORDER BY s.created_at DESC
	`, advisorID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Student
	for rows.Next() {
		var s models.Student

		err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.StudentID,
			&s.ProgramStudy,
			&s.AcademicYear,
			&s.AdvisorID,
			&s.FullName,
			&s.CreatedAt,
		)
		if err != nil {
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
		SET user_id=$1, student_id=$2, program_study=$3, academic_year=$4
		WHERE id = $5
	`, req.UserID, req.StudentID, req.ProgramStudy, req.AcademicYear, id)

	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, sql.ErrNoRows
	}

	return r.GetByID(id)
}

func (r *studentRepository) UpdateAdvisor(id string, advisorID string) error {
	result, err := r.db.Exec(`
		UPDATE students
		SET advisor_id = $1
		WHERE id = $2
	`, advisorID, id)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
