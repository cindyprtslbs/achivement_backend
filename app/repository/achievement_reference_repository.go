package repository

import (
	"database/sql"
	models "achievement_backend/app/model"
	"time"
)

type AchievementReferenceRepository interface {
	GetAll() ([]models.AchievementReference, error)
	GetByID(id string) (*models.AchievementReference, error)
	GetByStudentID(studentID string) ([]models.AchievementReference, error)

	Create(req models.CreateAchievementReferenceRequest) (*models.AchievementReference, error)
	Update(id string, req models.UpdateAchievementReferenceRequest) (*models.AchievementReference, error)

	Submit(id string) (*models.AchievementReference, error)
	Verify(id string, verifierID string) (*models.AchievementReference, error)
	Reject(id string, verifierID string, note string) (*models.AchievementReference, error)

	SoftDelete(id string, userID string) (*models.AchievementReference, error)
	Delete(id string) error
}

type achievementReferenceRepository struct {
	db *sql.DB
}

func NewAchievementReferenceRepository(db *sql.DB) AchievementReferenceRepository {
	return &achievementReferenceRepository{db: db}
}

func (r *achievementReferenceRepository) GetAll() ([]models.AchievementReference, error) {
	rows, err := r.db.Query(`
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.AchievementReference
	for rows.Next() {
		var a models.AchievementReference
		if err := rows.Scan(
			&a.ID,
			&a.StudentID,
			&a.MongoAchievementID,
			&a.Status,
			&a.SubmittedAt,
			&a.VerifiedAt,
			&a.VerifiedBy,
			&a.RejectionNote,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}

	return list, nil
}

func (r *achievementReferenceRepository) GetByID(id string) (*models.AchievementReference, error) {
	var a models.AchievementReference

	row := r.db.QueryRow(`
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE id = $1
	`, id)

	err := row.Scan(
		&a.ID,
		&a.StudentID,
		&a.MongoAchievementID,
		&a.Status,
		&a.SubmittedAt,
		&a.VerifiedAt,
		&a.VerifiedBy,
		&a.RejectionNote,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *achievementReferenceRepository) GetByStudentID(studentID string) ([]models.AchievementReference, error) {
	rows, err := r.db.Query(`
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE student_id = $1
		ORDER BY created_at DESC
	`, studentID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.AchievementReference
	for rows.Next() {
		var a models.AchievementReference
		if err := rows.Scan(
			&a.ID,
			&a.StudentID,
			&a.MongoAchievementID,
			&a.Status,
			&a.SubmittedAt,
			&a.VerifiedAt,
			&a.VerifiedBy,
			&a.RejectionNote,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, err
		}

		list = append(list, a)
	}

	return list, nil
}

func (r *achievementReferenceRepository) Create(req models.CreateAchievementReferenceRequest) (*models.AchievementReference, error) {
	var id string

	err := r.db.QueryRow(`
		INSERT INTO achievement_references (
			student_id, mongo_achievement_id, status,
			submitted_at, verified_at, verified_by,
			rejection_note, created_at, updated_at
		)
		VALUES ($1, $2, $3, NULL, NULL, NULL, NULL, $4, $5)
		RETURNING id
	`,
		req.StudentID,
		req.MongoAchievementID,
		req.Status,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *achievementReferenceRepository) Update(id string, req models.UpdateAchievementReferenceRequest) (*models.AchievementReference, error) {
	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET student_id=$1,
		    mongo_achievement_id=$2,
		    status=$3,
		    updated_at=$4
		WHERE id = $5
	`,
		req.StudentID,
		req.MongoAchievementID,
		req.Status,
		time.Now(),
		id,
	)
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *achievementReferenceRepository) Submit(id string) (*models.AchievementReference, error) {
	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='submitted',
		    submitted_at=$1,
		    updated_at=$1
		WHERE id=$2
	`, time.Now(), id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *achievementReferenceRepository) Verify(id string, verifierID string) (*models.AchievementReference, error) {
	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='verified',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=NULL,
		    updated_at=$1
		WHERE id=$3
	`, time.Now(), verifierID, id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *achievementReferenceRepository) Reject(id string, verifierID string, note string) (*models.AchievementReference, error) {
	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='rejected',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=$3,
		    updated_at=$1
		WHERE id=$4
	`, time.Now(), verifierID, note, id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *achievementReferenceRepository) SoftDelete(id string, userID string) (*models.AchievementReference, error) {
	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='deleted',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=NULL,
		    updated_at=$1
		WHERE id=$3
	`, time.Now(), userID, id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *achievementReferenceRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM achievement_references WHERE id=$1`, id)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
