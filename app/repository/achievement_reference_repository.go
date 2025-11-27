package repository

import (
	"database/sql"
	"time"
	models "achievement_backend/app/model"
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

	err := r.db.QueryRow(`
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE id = $1
	`, id).Scan(
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

// ================= CREATE REFERENCE (FR-003) =================

func (r *achievementReferenceRepository) Create(studentID string, mongoID string) (*models.AchievementReference, error) {
	var id string
	now := time.Now()

	err := r.db.QueryRow(`
		INSERT INTO achievement_references (
			student_id, mongo_achievement_id, status,
			submitted_at, verified_at, verified_by,
			rejection_note, created_at, updated_at
		) VALUES (
			$1, $2, 'submitted',
			$3, NULL, NULL,
			NULL, $3, $3
		)
		RETURNING id
	`,
		studentID,
		mongoID,
		now,
	).Scan(&id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

// ================= SUBMIT (FR-003) =================

func (r *achievementReferenceRepository) Submit(id string) error {
	now := time.Now()

	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='submitted',
		    submitted_at=$1,
		    updated_at=$1
		WHERE id=$2
	`, now, id)

	return err
}

// ================= VERIFY (FR-004) =================

func (r *achievementReferenceRepository) Verify(id string, verifierID string) error {
	now := time.Now()

	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='verified',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=NULL,
		    updated_at=$1
		WHERE id=$3
	`, now, verifierID, id)

	return err
}

// ================= REJECT (FR-004) =================

func (r *achievementReferenceRepository) Reject(id string, verifierID string, note string) error {
	now := time.Now()

	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='rejected',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=$3,
		    updated_at=$1
		WHERE id=$4
	`, now, verifierID, note, id)

	return err
}

// ================= SOFT DELETE (FR-010) =================

func (r *achievementReferenceRepository) SoftDelete(id string, userID string) error {
	now := time.Now()

	_, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='deleted',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=NULL,
		    updated_at=$1
		WHERE id=$3
	`, now, userID, id)

	return err
}
