package repository

import (
	"database/sql"
	models "achievement_backend/app/model"
	"time"
)

type LecturerRepository interface {
	GetAll() ([]models.Lecturer, error)
	GetByID(id string) (*models.Lecturer, error)
	GetByUserID(userID string) (*models.Lecturer, error)
	Create(req models.CreateLecturerRequest) (*models.Lecturer, error)
	Update(id string, req models.UpdateLecturerRequest) (*models.Lecturer, error)
	Delete(id string) error
}

type lecturerRepository struct {
	db *sql.DB
}

func NewLecturerRepository(db *sql.DB) LecturerRepository {
	return &lecturerRepository{db: db}
}

func (r *lecturerRepository) GetAll() ([]models.Lecturer, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Lecturer
	for rows.Next() {
		var l models.Lecturer
		if err := rows.Scan(
			&l.ID,
			&l.UserID,
			&l.LecturerID,
			&l.Department,
			&l.CreatedAt,
		); err != nil {
			return nil, err
		}

		list = append(list, l)
	}

	return list, nil
}

func (r *lecturerRepository) GetByID(id string) (*models.Lecturer, error) {
	var l models.Lecturer

	row := r.db.QueryRow(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		WHERE id = $1
	`, id)

	err := row.Scan(
		&l.ID,
		&l.UserID,
		&l.LecturerID,
		&l.Department,
		&l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (r *lecturerRepository) GetByUserID(userID string) (*models.Lecturer, error) {
	var l models.Lecturer

	row := r.db.QueryRow(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		WHERE user_id = $1
	`, userID)

	err := row.Scan(
		&l.ID,
		&l.UserID,
		&l.LecturerID,
		&l.Department,
		&l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (r *lecturerRepository) Create(req models.CreateLecturerRequest) (*models.Lecturer, error) {
	var id string

	err := r.db.QueryRow(`
		INSERT INTO lecturers (user_id, lecturer_id, department, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, req.UserID, req.LecturerID, req.Department, time.Now()).Scan(&id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *lecturerRepository) Update(id string, req models.UpdateLecturerRequest) (*models.Lecturer, error) {
	result, err := r.db.Exec(`
		UPDATE lecturers
		SET user_id=$1, lecturer_id=$2, department=$3
		WHERE id = $4
	`, req.UserID, req.LecturerID, req.Department, id)

	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, sql.ErrNoRows
	}

	return r.GetByID(id)
}

func (r *lecturerRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM lecturers WHERE id=$1`, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
