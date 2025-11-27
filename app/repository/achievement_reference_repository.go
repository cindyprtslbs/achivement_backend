package repository

import (
	models "achievement_backend/app/model"
	"database/sql"
	"log"
	"time"

	"github.com/lib/pq"
)

type AchievementReferenceRepository interface {
	GetAll() ([]models.AchievementReference, error)
	GetByID(id string) (*models.AchievementReference, error)
	GetByStudentID(studentID string) ([]models.AchievementReference, error)
	GetByMongoAchievementID(mongoID string) (*models.AchievementReference, error)
	GetByAdviseesWithPagination(studentIDs []string, limit int, offset int) ([]models.AchievementReference, int64, error)
	Create(studentID string, mongoID string) (*models.AchievementReference, error)
	Submit(id string) error
	Verify(id string, verifierID string) error
	Reject(id string, verifierID string, note string) error
	SoftDelete(id string) error
}

type achievementReferenceRepository struct {
	db *sql.DB
}

func NewAchievementReferenceRepository(db *sql.DB) AchievementReferenceRepository {
	return &achievementReferenceRepository{db: db}
}

// ================= GET ALL =================
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
			&a.ID, &a.StudentID, &a.MongoAchievementID, &a.Status,
			&a.SubmittedAt, &a.VerifiedAt, &a.VerifiedBy,
			&a.RejectionNote, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

// ================= GET BY ID =================
func (r *achievementReferenceRepository) GetByID(id string) (*models.AchievementReference, error) {
	var a models.AchievementReference
	err := r.db.QueryRow(`
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE id=$1
	`, id).Scan(
		&a.ID, &a.StudentID, &a.MongoAchievementID, &a.Status,
		&a.SubmittedAt, &a.VerifiedAt, &a.VerifiedBy,
		&a.RejectionNote, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ================= GET BY STUDENT =================
func (r *achievementReferenceRepository) GetByStudentID(studentID string) ([]models.AchievementReference, error) {
	rows, err := r.db.Query(`
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE student_id=$1
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
			&a.ID, &a.StudentID, &a.MongoAchievementID, &a.Status,
			&a.SubmittedAt, &a.VerifiedAt, &a.VerifiedBy,
			&a.RejectionNote, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

// ================= GET BY MONGO ACHIEVEMENT ID =================
func (r *achievementReferenceRepository) GetByMongoAchievementID(mongoID string) (*models.AchievementReference, error) {
	var a models.AchievementReference

	log.Printf("[REPO-GET-BY-MONGO] Querying achievement_references where mongo_achievement_id='%s'", mongoID)

	err := r.db.QueryRow(`
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE mongo_achievement_id=$1
	`, mongoID).Scan(
		&a.ID, &a.StudentID, &a.MongoAchievementID, &a.Status,
		&a.SubmittedAt, &a.VerifiedAt, &a.VerifiedBy,
		&a.RejectionNote, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		log.Printf("[REPO-GET-BY-MONGO] Error querying: %v", err)
		return nil, err
	}

	log.Printf("[REPO-GET-BY-MONGO] Found reference: %+v", a)
	return &a, nil
}

// ================= CREATE =================
func (r *achievementReferenceRepository) Create(studentID string, mongoID string) (*models.AchievementReference, error) {
	var id string
	now := time.Now()

	log.Printf("[REPO-CREATE] Attempting to create reference - studentID: %s, mongoID: %s", studentID, mongoID)

	err := r.db.QueryRow(`
		INSERT INTO achievement_references (
			student_id, mongo_achievement_id, status,
			created_at, updated_at
		) VALUES (
			$1, $2, 'draft', $3, $3
		) RETURNING id
	`, studentID, mongoID, now).Scan(&id)
	if err != nil {
		log.Printf("[REPO-CREATE] Error inserting reference: %v", err)
		return nil, err
	}

	log.Printf("[REPO-CREATE] Successfully created reference with ID: %s", id)
	return r.GetByID(id)
}

// ================= SUBMIT =================
func (r *achievementReferenceRepository) Submit(id string) error {
	now := time.Now()

	res, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='submitted',
		    submitted_at=$1,
		    updated_at=$1
		WHERE id=$2
		  AND status='draft'
	`, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ================= VERIFY =================
func (r *achievementReferenceRepository) Verify(id string, verifierID string) error {
	now := time.Now()
	res, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='verified',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=NULL,
		    updated_at=$1
		WHERE id=$3 AND status='submitted'
	`, now, verifierID, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ================= REJECT =================
func (r *achievementReferenceRepository) Reject(id string, verifierID string, note string) error {
	now := time.Now()
	res, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='rejected',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=$3,
		    updated_at=$1
		WHERE id=$4 AND status='submitted'
	`, now, verifierID, note, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ================= SOFT DELETE =================
func (r *achievementReferenceRepository) SoftDelete(id string) error {
	now := time.Now()
	res, err := r.db.Exec(`
		UPDATE achievement_references
		SET status='deleted',
		    updated_at=$1
		WHERE id=$2 AND status='draft'
	`, now, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ================= GET BY ADVISEES WITH PAGINATION =================
func (r *achievementReferenceRepository) GetByAdviseesWithPagination(studentIDs []string, limit int, offset int) ([]models.AchievementReference, int64, error) {
	if len(studentIDs) == 0 {
		return []models.AchievementReference{}, 0, nil
	}

	query := `
SELECT id, student_id, mongo_achievement_id, status,
       submitted_at, verified_at, verified_by,
       rejection_note, created_at, updated_at
FROM achievement_references
WHERE student_id = ANY($1::uuid[])
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`

	rows, err := r.db.Query(query, pq.Array(studentIDs), limit, offset)
	if err != nil {
		log.Printf("[REPO] Error querying achievements: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var list []models.AchievementReference
	for rows.Next() {
		var a models.AchievementReference
		if err := rows.Scan(
			&a.ID, &a.StudentID, &a.MongoAchievementID, &a.Status,
			&a.SubmittedAt, &a.VerifiedAt, &a.VerifiedBy,
			&a.RejectionNote, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			log.Printf("[REPO] Error scanning row: %v", err)
			return nil, 0, err
		}
		list = append(list, a)
	}

	countQuery := `
SELECT COUNT(*)
FROM achievement_references
WHERE student_id = ANY($1::uuid[])
`
	var total int64
	err = r.db.QueryRow(countQuery, pq.Array(studentIDs)).Scan(&total)
	if err != nil {
		log.Printf("[REPO] Error counting rows: %v", err)
		return nil, 0, err
	}

	log.Printf("[REPO] Retrieved %d achievements for advisees", len(list))
	return list, total, nil
}
