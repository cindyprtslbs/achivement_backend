package repository

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	// models "achievement_backend/app/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupAchievementRefRepo(t *testing.T) (*sql.DB, sqlmock.Sqlmock, AchievementReferenceRepository) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	repo := NewAchievementReferenceRepository(db)
	return db, mock, repo
}

func TestAchievementReference_GetByID_Success(t *testing.T) {
	db, mock, repo := setupAchievementRefRepo(t)
	defer db.Close()

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "student_id", "mongo_achievement_id", "status",
		"submitted_at", "verified_at", "verified_by",
		"rejection_note", "created_at", "updated_at",
	}).AddRow(
		"1", "434231016", "2345678909876543256", "draft",
		nil, nil, nil,
		nil, now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE id=$1
	`)).WithArgs("1").WillReturnRows(rows)

	res, err := repo.GetByID("1")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "draft", res.Status)
}

func TestAchievementReference_Create(t *testing.T) {
	db, mock, repo := setupAchievementRefRepo(t)
	defer db.Close()

	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO achievement_references (
			student_id, mongo_achievement_id, status,
			created_at, updated_at
		) VALUES (
			$1, $2, 'draft', $3, $3
		) RETURNING id
	`)).
		WithArgs("434231016", "2345678909876543256", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	// GetByID dipanggil setelah insert
	mock.ExpectQuery("SELECT id, student_id").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "student_id", "mongo_achievement_id", "status",
			"submitted_at", "verified_at", "verified_by",
			"rejection_note", "created_at", "updated_at",
		}).AddRow(
			"123456sd-jhgf34567-jhg45678", "434231016", "2345678909876543256", "draft",
			nil, nil, nil, nil, now, now,
		))

	res, err := repo.Create("434231016", "2345678909876543256")

	assert.NoError(t, err)
	assert.Equal(t, "draft", res.Status)
}

func TestAchievementReference_Submit(t *testing.T) {
	db, mock, repo := setupAchievementRefRepo(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE achievement_references
		SET status='submitted',
		    submitted_at=$1,
		    updated_at=$1
		WHERE id=$2
		  AND status='draft'
	`)).
		WithArgs(sqlmock.AnyArg(), "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Submit("1")

	assert.NoError(t, err)
}

func TestAchievementReference_Verify(t *testing.T) {
	db, mock, repo := setupAchievementRefRepo(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE achievement_references
		SET status='verified',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=NULL,
		    updated_at=$1
		WHERE id=$3 AND status='submitted'
	`)).
		WithArgs(sqlmock.AnyArg(), "DSN002", "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Verify("1", "DSN002")

	assert.NoError(t, err)
}

func TestAchievementReference_Reject(t *testing.T) {
	db, mock, repo := setupAchievementRefRepo(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE achievement_references
		SET status='rejected',
		    verified_at=$1,
		    verified_by=$2,
		    rejection_note=$3,
		    updated_at=$1
		WHERE id=$4 AND status='submitted'
	`)).
		WithArgs(sqlmock.AnyArg(), "DSN002", "note", "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Reject("1", "DSN002", "note")

	assert.NoError(t, err)
}

func TestAchievementReference_SoftDelete(t *testing.T) {
	db, mock, repo := setupAchievementRefRepo(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE achievement_references
		SET status='deleted',
		    updated_at=$1
		WHERE id=$2
		  AND status='draft'
	`)).
		WithArgs(sqlmock.AnyArg(), "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SoftDelete("1")

	assert.NoError(t, err)
}
