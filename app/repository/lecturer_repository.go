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
	GetByLecturerID(lecturerID string) (*models.Lecturer, error)
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

	if err == sql.ErrNoRows {
		return nil, nil // ← FIX
	}

	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (r *lecturerRepository) Create(req models.CreateLecturerRequest) (*models.Lecturer, error) {
	var id string
	var createdAt time.Time

	err := r.db.QueryRow(`
		INSERT INTO lecturers (user_id, lecturer_id, department)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, req.UserID, req.LecturerID, req.Department).Scan(&id, &createdAt)
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *lecturerRepository) Update(id string, req models.UpdateLecturerRequest) (*models.Lecturer, error) {
	_, err := r.db.Exec(`
		UPDATE lecturers
		SET lecturer_id=$1, department=$2
		WHERE id = $3
	`, req.LecturerID, req.Department, id)

	if err != nil {
		return nil, err
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

func (r *lecturerRepository) GetByLecturerID(lecturerID string) (*models.Lecturer, error) {
	var l models.Lecturer

	row := r.db.QueryRow(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		WHERE lecturer_id = $1
	`, lecturerID)

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
