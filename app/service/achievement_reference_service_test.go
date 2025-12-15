package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	models "achievement_backend/app/model"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

//
// =======================================================
// MOCK AchievementReferenceRepository
// =======================================================
//

type mockAchievementRefRepo struct {
	ref *models.AchievementReference
}

func (m *mockAchievementRefRepo) GetAll() ([]models.AchievementReference, error) {
	return nil, nil
}

func (m *mockAchievementRefRepo) GetAllWithPagination(limit, offset int) ([]models.AchievementReference, int64, error) {
	return []models.AchievementReference{*m.ref}, 1, nil
}

func (m *mockAchievementRefRepo) GetByID(id string) (*models.AchievementReference, error) {
	return m.ref, nil
}

func (m *mockAchievementRefRepo) GetByStudentID(studentID string) ([]models.AchievementReference, error) {
	return []models.AchievementReference{*m.ref}, nil
}

func (m *mockAchievementRefRepo) GetByMongoAchievementID(mongoID string) (*models.AchievementReference, error) {
	return m.ref, nil
}

func (m *mockAchievementRefRepo) GetByAdviseesWithPagination(ids []string, limit, offset int) ([]models.AchievementReference, int64, error) {
	return nil, 0, nil
}

func (m *mockAchievementRefRepo) Create(studentID, mongoID string) (*models.AchievementReference, error) {
	return m.ref, nil
}

func (m *mockAchievementRefRepo) Submit(id string) error {
	now := time.Now()
	m.ref.Status = models.StatusSubmitted
	m.ref.SubmittedAt = &now
	return nil
}

func (m *mockAchievementRefRepo) Verify(id, verifierID string) error {
	now := time.Now()
	m.ref.Status = models.StatusVerified
	m.ref.VerifiedAt = &now
	m.ref.VerifiedBy = &verifierID
	return nil
}

func (m *mockAchievementRefRepo) Reject(id, verifierID, note string) error {
	m.ref.Status = models.StatusRejected
	m.ref.RejectionNote = &note
	return nil
}

func (m *mockAchievementRefRepo) SoftDelete(id string) error {
	return nil
}

//
// =======================================================
// MOCK MongoAchievementRepository (WAJIB LENGKAP)
// =======================================================
//

type mockMongoAchievementRepo struct{}

func (m *mockMongoAchievementRepo) GetAll(ctx context.Context) ([]models.Achievement, error) {
	return nil, nil
}

func (m *mockMongoAchievementRepo) GetByAdvisor(ctx context.Context, studentIDs []string) ([]models.Achievement, error) {
	return nil, nil
}

func (m *mockMongoAchievementRepo) CreateDraft(ctx context.Context, studentID string, req *models.CreateAchievementRequest, points int) (*models.Achievement, error) {
	return nil, nil
}

func (m *mockMongoAchievementRepo) GetByID(ctx context.Context, id string) (*models.Achievement, error) {
	return &models.Achievement{Title: "Mock Achievement"}, nil
}

func (m *mockMongoAchievementRepo) GetByStudentID(ctx context.Context, studentID string) ([]models.Achievement, error) {
	return nil, nil
}

func (m *mockMongoAchievementRepo) UpdateDraft(ctx context.Context, id string, req *models.UpdateAchievementRequest, points int) (*models.Achievement, error) {
	return nil, nil
}

func (m *mockMongoAchievementRepo) UpdateAttachments(ctx context.Context, id string, attachments []models.Attachment) (*models.Achievement, error) {
	return nil, nil
}

func (m *mockMongoAchievementRepo) SoftDelete(ctx context.Context, id string) error {
	return nil
}

func (m *mockMongoAchievementRepo) GetManyByIDs(ctx context.Context, ids []string) (map[string]models.Achievement, error) {
	result := map[string]models.Achievement{}
	for _, id := range ids {
		result[id] = models.Achievement{Title: "Mock Achievement"}
	}
	return result, nil
}

func (m *mockMongoAchievementRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	return nil
}

//
// =======================================================
// MOCK StudentRepository (NAMA UNIK)
// =======================================================
//

type mockAchievementStudentRepo struct{}

func (m *mockAchievementStudentRepo) GetAll() ([]models.Student, error) { return nil, nil }
func (m *mockAchievementStudentRepo) GetByID(id string) (*models.Student, error) {
	return &models.Student{ID: id}, nil
}
func (m *mockAchievementStudentRepo) GetByStudentID(studentID string) (*models.Student, error) {
	return nil, nil
}
func (m *mockAchievementStudentRepo) GetByUserID(userID string) (*models.Student, error) {
	return &models.Student{ID: "student-1"}, nil
}
func (m *mockAchievementStudentRepo) GetByAdvisorID(advisorID string) ([]models.Student, error) {
	return nil, nil
}
func (m *mockAchievementStudentRepo) Create(req models.CreateStudentRequest) (*models.Student, error) {
	return nil, nil
}
func (m *mockAchievementStudentRepo) Update(id string, req models.UpdateStudentRequest) (*models.Student, error) {
	return nil, nil
}
func (m *mockAchievementStudentRepo) UpdateAdvisor(id string, advisorID string) error {
	return nil
}

//
// =======================================================
// MOCK LecturerRepository (NAMA UNIK)
// =======================================================
//

type mockAchievementLecturerRepo struct{}

func (m *mockAchievementLecturerRepo) GetAll() ([]models.Lecturer, error) { return nil, nil }
func (m *mockAchievementLecturerRepo) GetByID(id string) (*models.Lecturer, error) {
	return nil, nil
}
func (m *mockAchievementLecturerRepo) GetByUserID(userID string) (*models.Lecturer, error) {
	return &models.Lecturer{ID: "lecturer-1"}, nil
}
func (m *mockAchievementLecturerRepo) GetByLecturerID(id string) (*models.Lecturer, error) {
	return nil, nil
}
func (m *mockAchievementLecturerRepo) Create(req models.CreateLecturerRequest) (*models.Lecturer, error) {
	return nil, nil
}
func (m *mockAchievementLecturerRepo) Update(id string, req models.UpdateLecturerRequest) (*models.Lecturer, error) {
	return nil, nil
}

//
// =======================================================
// SETUP
// =======================================================
//

func setupAchievementService() (*fiber.App, *mockAchievementRefRepo) {
	app := fiber.New()

	ref := &models.AchievementReference{
		ID:                 "ref-1",
		StudentID:          "student-1",
		MongoAchievementID: "mongo-1",
		Status:             models.StatusDraft,
	}

	refRepo := &mockAchievementRefRepo{ref: ref}

	service := NewAchievementReferenceService(
		refRepo,
		&mockMongoAchievementRepo{},
		&mockAchievementStudentRepo{},
		&mockAchievementLecturerRepo{},
	)

	app.Get("/refs", service.GetAll)
	app.Get("/refs/:id", service.GetByID)
	app.Post("/refs/:id/submit", func(c *fiber.Ctx) error {
		c.Locals("role_name", "Mahasiswa")
		c.Locals("user_id", "user-1")
		return service.Submit(c)
	})
	app.Post("/refs/:id/verify", func(c *fiber.Ctx) error {
		c.Locals("role_name", "Admin")
		c.Locals("user_id", "lecturer-1")
		return service.Verify(c)
	})
	app.Post("/refs/:id/reject", func(c *fiber.Ctx) error {
		c.Locals("role_name", "Admin")
		c.Locals("user_id", "lecturer-1")
		return service.Reject(c)
	})

	return app, refRepo
}

//
// =======================================================
// TESTS
// =======================================================
//

func TestGetAll(t *testing.T) {
	app, _ := setupAchievementService()
	req := httptest.NewRequest(http.MethodGet, "/refs", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSubmit(t *testing.T) {
	app, repo := setupAchievementService()
	req := httptest.NewRequest(http.MethodPost, "/refs/mongo-1/submit", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, models.StatusSubmitted, repo.ref.Status)
}

func TestVerify(t *testing.T) {
	app, repo := setupAchievementService()
	req := httptest.NewRequest(http.MethodPost, "/refs/mongo-1/verify", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, models.StatusVerified, repo.ref.Status)
}

func TestReject(t *testing.T) {
	app, repo := setupAchievementService()

	body, _ := json.Marshal(map[string]string{
		"rejection_note": "invalid data",
	})

	req := httptest.NewRequest(http.MethodPost, "/refs/mongo-1/reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, models.StatusRejected, repo.ref.Status)
}
