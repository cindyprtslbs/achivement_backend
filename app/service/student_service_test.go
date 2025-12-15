package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	models "achievement_backend/app/model"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type MockStudentRepo struct {
	GetAllFn       func() ([]models.Student, error)
	GetByIDFn      func(string) (*models.Student, error)
	CreateFn       func(models.CreateStudentRequest) (*models.Student, error)
	UpdateFn       func(string, models.UpdateStudentRequest) (*models.Student, error)
	UpdateAdvisorFn func(string, string) error
}

func (m *MockStudentRepo) GetAll() ([]models.Student, error) {
	return m.GetAllFn()
}
func (m *MockStudentRepo) GetByID(id string) (*models.Student, error) {
	return m.GetByIDFn(id)
}
func (m *MockStudentRepo) Create(req models.CreateStudentRequest) (*models.Student, error) {
	return m.CreateFn(req)
}
func (m *MockStudentRepo) Update(id string, req models.UpdateStudentRequest) (*models.Student, error) {
	return m.UpdateFn(id, req)
}
func (m *MockStudentRepo) UpdateAdvisor(id string, advisorID string) error {
	return m.UpdateAdvisorFn(id, advisorID)
}

/* unused methods (biar satisfy interface) */
func (m *MockStudentRepo) GetByStudentID(string) (*models.Student, error) { return nil, nil }
func (m *MockStudentRepo) GetByUserID(string) (*models.Student, error)    { return nil, nil }
func (m *MockStudentRepo) GetByAdvisorID(string) ([]models.Student, error) {
	return nil, nil
}

func TestStudentService_GetAll_Admin(t *testing.T) {
	mockRepo := &MockStudentRepo{
		GetAllFn: func() ([]models.Student, error) {
			return []models.Student{
				{ID: "1", StudentID: "2023"},
			}, nil
		},
	}

	app := fiber.New()
	service := NewStudentService(mockRepo, nil, nil)

	app.Get("/students", func(c *fiber.Ctx) error {
		c.Locals("role_name", "Admin")
		return service.GetAll(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/students", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestStudentService_GetAll_NonAdmin(t *testing.T) {
	mockRepo := &MockStudentRepo{}

	app := fiber.New()
	service := NewStudentService(mockRepo, nil, nil)

	app.Get("/students", func(c *fiber.Ctx) error {
		c.Locals("role_name", "Mahasiswa")
		return service.GetAll(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/students", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 403, resp.StatusCode)
}

func TestStudentService_Create(t *testing.T) {
	mockRepo := &MockStudentRepo{
		CreateFn: func(req models.CreateStudentRequest) (*models.Student, error) {
			return &models.Student{
				ID:        "1",
				StudentID: req.StudentID,
			}, nil
		},
	}

	app := fiber.New()
	service := NewStudentService(mockRepo, nil, nil)

	app.Post("/students", service.Create)

	body, _ := json.Marshal(models.CreateStudentRequest{
		UserID:       "user-1",
		StudentID:    "20231234",
		ProgramStudy: "Informatika",
		AcademicYear: "2023",
	})

	req := httptest.NewRequest(http.MethodPost, "/students", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, 201, resp.StatusCode)
}

func TestStudentService_UpdateAdvisor_Admin(t *testing.T) {
	mockRepo := &MockStudentRepo{
		GetByIDFn: func(id string) (*models.Student, error) {
			return &models.Student{ID: id}, nil
		},
		UpdateAdvisorFn: func(id, advisorID string) error {
			return nil
		},
	}

	app := fiber.New()
	service := NewStudentService(mockRepo, nil, nil)

	app.Put("/students/:id/advisor", func(c *fiber.Ctx) error {
		c.Locals("role_name", "Admin")
		return service.UpdateAdvisor(c)
	})

	body := []byte(`{"advisor_id":"advisor-1"}`)
	req := httptest.NewRequest(http.MethodPut, "/students/1/advisor", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}

