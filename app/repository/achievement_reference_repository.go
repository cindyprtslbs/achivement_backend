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

	Create(studentID string, mongoID string) (*models.AchievementReference, error)

	Submit(id string) error
	Verify(id string, verifierID string) error
	Reject(id string, verifierID string, note string) error
	SoftDelete(id string, userID string) error
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

func (r *achievementReferenceRepository) Create(studentID string, mongoID string) (*models.AchievementReference, error) {
	var id string

	err := r.db.QueryRow(`
		INSERT INTO achievement_references (
			student_id, mongo_achievement_id, status,
			submitted_at, verified_at, verified_by,
			rejection_note, created_at, updated_at
		)
		VALUES ($1, $2, 'submitted', NOW(), NULL, NULL, NULL, NOW(), NOW())
		RETURNING id
	`,
		studentID,
		mongoID,
	).Scan(&id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *achievementReferenceRepository) Submit(id string) error {
	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='submitted',
		    submitted_at=$1,
		    updated_at=$1
		WHERE id=$2
	`, time.Now(), id)

	return err
}

func (r *achievementReferenceRepository) Verify(id string, verifierID string) error {
	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='verified',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=NULL,
		    updated_at=$1
		WHERE id=$3
	`, time.Now(), verifierID, id)

	return err
}

func (r *achievementReferenceRepository) Reject(id string, verifierID string, note string) error {
	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='rejected',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=$3,
		    updated_at=$1
		WHERE id=$4
	`, time.Now(), verifierID, note, id)

	return err
}

func (r *achievementReferenceRepository) SoftDelete(id string, userID string) error {
	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='deleted',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=NULL,
		    updated_at=$1
		WHERE id=$3
	`, time.Now(), userID, id) 

	return err
}
