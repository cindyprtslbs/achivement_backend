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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//
// =======================================================
// MOCK MongoAchievementRepository (UNIK & LENGKAP)
// =======================================================
//

type mockAchMongoRepo struct {
	item *models.Achievement
}

func (m *mockAchMongoRepo) GetAll(ctx context.Context) ([]models.Achievement, error) {
	return []models.Achievement{*m.item}, nil
}

func (m *mockAchMongoRepo) GetByAdvisor(ctx context.Context, ids []string) ([]models.Achievement, error) {
	return []models.Achievement{*m.item}, nil
}

func (m *mockAchMongoRepo) CreateDraft(
	ctx context.Context,
	studentID string,
	req *models.CreateAchievementRequest,
	points int,
) (*models.Achievement, error) {

	mongoID := primitive.NewObjectID()
	p := float64(points)
	m.item = &models.Achievement{
		ID:        mongoID,
		StudentID: studentID,
		Title:     req.Title,
		Status:    models.StatusDraft,
		Points:    &p,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return m.item, nil
}

func (m *mockAchMongoRepo) GetByID(ctx context.Context, id string) (*models.Achievement, error) {
	return m.item, nil
}

func (m *mockAchMongoRepo) GetByStudentID(ctx context.Context, studentID string) ([]models.Achievement, error) {
	return []models.Achievement{*m.item}, nil
}

func (m *mockAchMongoRepo) UpdateDraft(
	ctx context.Context,
	id string,
	req *models.UpdateAchievementRequest,
	points int,
) (*models.Achievement, error) {

	p := float64(points)
	m.item.Title = req.Title
	m.item.Points = &p
	m.item.UpdatedAt = time.Now()
	return m.item, nil
}

func (m *mockAchMongoRepo) UpdateAttachments(
	ctx context.Context,
	id string,
	attachments []models.Attachment,
) (*models.Achievement, error) {

	m.item.Attachments = attachments
	return m.item, nil
}

func (m *mockAchMongoRepo) SoftDelete(ctx context.Context, id string) error {
	m.item.Status = models.StatusDeleted
	m.item.IsDeleted = true
	return nil
}

func (m *mockAchMongoRepo) GetManyByIDs(
	ctx context.Context,
	ids []string,
) (map[string]models.Achievement, error) {

	return map[string]models.Achievement{
		m.item.ID.Hex(): *m.item,
	}, nil
}

func (m *mockAchMongoRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	m.item.Status = status
	return nil
}

//
// =======================================================
// MOCK AchievementReferenceRepository (MINIMAL)
// =======================================================
//

type mockAchRefRepo struct{}

func (m *mockAchRefRepo) GetAll() ([]models.AchievementReference, error) {
	return nil, nil
}

func (m *mockAchRefRepo) GetAllWithPagination(limit, offset int) ([]models.AchievementReference, int64, error) {
	return nil, 0, nil
}

func (m *mockAchRefRepo) GetByID(id string) (*models.AchievementReference, error) {
	return nil, nil
}

func (m *mockAchRefRepo) GetByStudentID(studentID string) ([]models.AchievementReference, error) {
	return nil, nil
}

func (m *mockAchRefRepo) GetByMongoAchievementID(id string) (*models.AchievementReference, error) {
	return nil, nil
}

func (m *mockAchRefRepo) GetByAdviseesWithPagination(ids []string, limit, offset int) ([]models.AchievementReference, int64, error) {
	return nil, 0, nil
}

func (m *mockAchRefRepo) Create(studentID, mongoID string) (*models.AchievementReference, error) {
	return nil, nil
}

func (m *mockAchRefRepo) Submit(id string) error { return nil }
func (m *mockAchRefRepo) Verify(id, vid string) error {
	return nil
}
func (m *mockAchRefRepo) Reject(id, vid, note string) error {
	return nil
}
func (m *mockAchRefRepo) SoftDelete(id string) error { return nil }

//
// =======================================================
// MOCK StudentRepository (UNIK)
// =======================================================
//

type mockAchStudentRepo struct{}

func (m *mockAchStudentRepo) GetAll() ([]models.Student, error) { return nil, nil }
func (m *mockAchStudentRepo) GetByID(id string) (*models.Student, error) {
	return &models.Student{
		ID:        id,
		AdvisorID: ptTr("lecturer-1"),
	}, nil
}
func (m *mockAchStudentRepo) GetByStudentID(studentID string) (*models.Student, error) {
	return nil, nil
}
func (m *mockAchStudentRepo) GetByUserID(userID string) (*models.Student, error) {
	return &models.Student{ID: "student-1"}, nil
}
func (m *mockAchStudentRepo) GetByAdvisorID(advisorID string) ([]models.Student, error) {
	return []models.Student{{ID: "student-1"}}, nil
}
func (m *mockAchStudentRepo) Create(req models.CreateStudentRequest) (*models.Student, error) {
	return nil, nil
}
func (m *mockAchStudentRepo) Update(id string, req models.UpdateStudentRequest) (*models.Student, error) {
	return nil, nil
}
func (m *mockAchStudentRepo) UpdateAdvisor(id string, advisorID string) error {
	return nil
}

//
// =======================================================
// MOCK LecturerRepository (UNIK)
// =======================================================
//

type mockAchLecturerRepo struct{}

func (m *mockAchLecturerRepo) GetAll() ([]models.Lecturer, error) { return nil, nil }
func (m *mockAchLecturerRepo) GetByID(id string) (*models.Lecturer, error) {
	return nil, nil
}
func (m *mockAchLecturerRepo) GetByUserID(userID string) (*models.Lecturer, error) {
	return &models.Lecturer{ID: "lecturer-1"}, nil
}
func (m *mockAchLecturerRepo) GetByLecturerID(id string) (*models.Lecturer, error) {
	return nil, nil
}
func (m *mockAchLecturerRepo) Create(req models.CreateLecturerRequest) (*models.Lecturer, error) {
	return nil, nil
}
func (m *mockAchLecturerRepo) Update(id string, req models.UpdateLecturerRequest) (*models.Lecturer, error) {
	return nil, nil
}

//
// =======================================================
// TEST UTIL
// =======================================================
//

func ptTr[T any](v T) *T { return &v }

//
// =======================================================
// TEST: CREATE DRAFT
// =======================================================
//

func TestAchievementMongo_CreateDraft(t *testing.T) {
	app := fiber.New()

	mongoRepo := &mockAchMongoRepo{}

	service := NewAchievementMongoService(
		mongoRepo,
		&mockAchRefRepo{},
		&mockAchStudentRepo{},
		&mockAchLecturerRepo{},
	)

	app.Post("/api/v1/achievements", func(c *fiber.Ctx) error {
		c.Locals("role_name", "Mahasiswa")
		c.Locals("user_id", "user-1")
		return service.CreateDraft(c)
	})

	body := models.CreateAchievementRequest{
		AchievementType: "competition",
		Title:           "Lomba Nasional",
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/achievements", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	assert.Equal(t, "Lomba Nasional", mongoRepo.item.Title)
}

//
// =======================================================
// TEST: UPDATE DRAFT
// =======================================================
//

func TestAchievementMongo_UpdateDraft(t *testing.T) {
	app := fiber.New()

	mongoID := primitive.NewObjectID()
	mongoRepo := &mockAchMongoRepo{
		item: &models.Achievement{
			ID:        mongoID,
			StudentID: "student-1",
			Status:    models.StatusDraft,
		},
	}

	service := NewAchievementMongoService(
		mongoRepo,
		&mockAchRefRepo{},
		&mockAchStudentRepo{},
		&mockAchLecturerRepo{},
	)

	app.Put("/api/v1/achievements/:id", func(c *fiber.Ctx) error {
		c.Locals("role_name", "Mahasiswa")
		c.Locals("user_id", "user-1")
		return service.UpdateDraft(c)
	})

	body := models.UpdateAchievementRequest{Title: "Updated Title"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/achievements/"+mongoID.Hex(),
		bytes.NewReader(b),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "Updated Title", mongoRepo.item.Title)
}

//
// =======================================================
// TEST: SOFT DELETE
// =======================================================
//

func TestAchievementMongo_SoftDelete(t *testing.T) {
	app := fiber.New()

	mongoID := primitive.NewObjectID()

	mongoRepo := &mockAchMongoRepo{
		item: &models.Achievement{
			ID:        mongoID,
			StudentID: "student-1",
			Status:    models.StatusDraft,
		},
	}

	service := NewAchievementMongoService(
		mongoRepo,
		&mockAchRefRepo{},
		&mockAchStudentRepo{},
		&mockAchLecturerRepo{},
	)

	app.Delete("/api/v1/achievements/:id", func(c *fiber.Ctx) error {
		c.Locals("role_name", "Mahasiswa")
		c.Locals("user_id", "user-1")
		return service.SoftDelete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/achievements/mongo-1", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, models.StatusDeleted, mongoRepo.item.Status)
}
