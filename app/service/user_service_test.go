package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	models "achievement_backend/app/model"
	// "achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

//
// =======================================================
// MOCK USER REPOSITORY
// =======================================================
//

type mockUserRepo struct {
	data map[string]*models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{data: map[string]*models.User{}}
}

func (m *mockUserRepo) GetAll() ([]models.User, error) {
	var res []models.User
	for _, u := range m.data {
		res = append(res, *u)
	}
	return res, nil
}

func (m *mockUserRepo) GetByID(id string) (*models.User, error) {
	u, ok := m.data[id]
	if !ok {
		return nil, fiber.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByEmail(email string) (*models.User, error) {
	for _, u := range m.data {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) GetByUsername(username string) (*models.User, error) {
	for _, u := range m.data {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) Create(req models.CreateUserRequest) (*models.User, error) {
	u := &models.User{
		ID:       "1",
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
		IsActive: true,
	}
	m.data[u.ID] = u
	return u, nil
}

func (m *mockUserRepo) UpdatePartial(u *models.User) (*models.User, error) {
	m.data[u.ID] = u
	return u, nil
}

func (m *mockUserRepo) UpdatePassword(id string, passwordHash string) error {
	u, ok := m.data[id]
	if !ok {
		return fiber.ErrNotFound
	}
	u.PasswordHash = passwordHash
	return nil
}

func (m *mockUserRepo) Delete(id string) error {
	if _, ok := m.data[id]; !ok {
		return fiber.ErrNotFound
	}
	delete(m.data, id)
	return nil
}

//
// =======================================================
// MOCK ROLE REPOSITORY (FULL)
// =======================================================
//

type mockRoleRepo struct{}

func (m *mockRoleRepo) GetAll() ([]models.Role, error) {
	return []models.Role{}, nil
}

func (m *mockRoleRepo) GetByID(id string) (*models.Role, error) {
	return &models.Role{ID: id, Name: "Mahasiswa"}, nil
}

func (m *mockRoleRepo) Create(req models.CreateRoleRequest) (*models.Role, error) {
	return &models.Role{ID: "1", Name: req.Name}, nil
}

func (m *mockRoleRepo) Update(id string, req models.UpdateRoleRequest) (*models.Role, error) {
	return &models.Role{ID: id, Name: req.Name}, nil
}

func (m *mockRoleRepo) Delete(id string) error {
	return nil
}

//
// =======================================================
// MOCK STUDENT REPOSITORY (FULL INTERFACE)
// =======================================================
//

type mockStudentRepo struct{}

func (m *mockStudentRepo) GetAll() ([]models.Student, error) {
	return []models.Student{}, nil
}

func (m *mockStudentRepo) GetByID(id string) (*models.Student, error) {
	return &models.Student{ID: id}, nil
}

func (m *mockStudentRepo) GetByStudentID(studentID string) (*models.Student, error) {
	return nil, nil
}

func (m *mockStudentRepo) GetByUserID(userID string) (*models.Student, error) {
	return nil, nil
}

func (m *mockStudentRepo) GetByAdvisorID(advisorID string) ([]models.Student, error) {
	return []models.Student{}, nil
}

func (m *mockStudentRepo) Create(req models.CreateStudentRequest) (*models.Student, error) {
	return &models.Student{ID: "1", UserID: req.UserID}, nil
}

func (m *mockStudentRepo) Update(id string, req models.UpdateStudentRequest) (*models.Student, error) {
	return &models.Student{ID: id, UserID: req.UserID}, nil
}

func (m *mockStudentRepo) UpdateAdvisor(id string, advisorID string) error {
	return nil
}

//
// =======================================================
// MOCK LECTURER REPOSITORY (FULL INTERFACE)
// =======================================================
//

type mockLecturerRepo struct{}

func (m *mockLecturerRepo) GetAll() ([]models.Lecturer, error) {
	return []models.Lecturer{}, nil
}

func (m *mockLecturerRepo) GetByID(id string) (*models.Lecturer, error) {
	return &models.Lecturer{ID: id}, nil
}

func (m *mockLecturerRepo) GetByUserID(userID string) (*models.Lecturer, error) {
	return nil, nil
}

func (m *mockLecturerRepo) GetByLecturerID(lecturerID string) (*models.Lecturer, error) {
	return nil, nil
}

func (m *mockLecturerRepo) Create(req models.CreateLecturerRequest) (*models.Lecturer, error) {
	return &models.Lecturer{ID: "1", UserID: req.UserID}, nil
}

func (m *mockLecturerRepo) Update(id string, req models.UpdateLecturerRequest) (*models.Lecturer, error) {
	return &models.Lecturer{ID: id}, nil
}

//
// =======================================================
// SETUP SERVICE
// =======================================================
//

func setupUserService() (*fiber.App, *mockUserRepo) {
	app := fiber.New()

	userRepo := newMockUserRepo()

	service := NewUserService(
		userRepo,
		&mockRoleRepo{},
		&mockStudentRepo{},
		&mockLecturerRepo{},
	)

	app.Get("/users", service.GetAll)
	app.Get("/users/:id", service.GetByID)
	app.Post("/users", service.Create)
	app.Put("/users/:id/password", service.UpdatePassword)
	app.Delete("/users/:id", service.Delete)

	return app, userRepo
}

//
// =======================================================
// TEST CASES
// =======================================================
//

func TestUserService_GetAll(t *testing.T) {
	app, repo := setupUserService()

	repo.data["1"] = &models.User{ID: "1", Username: "cindy"}
	repo.data["2"] = &models.User{ID: "2", Username: "jane"}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestUserService_Create(t *testing.T) {
	app, _ := setupUserService()

	body := models.CreateUserRequest{
		Username:     "cindy",
		Email:        "cindy@mail.com",
		PasswordHash: "password",
		FullName:     "cindy Doe",
	}

	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestUserService_UpdatePassword(t *testing.T) {
	app, repo := setupUserService()

	repo.data["1"] = &models.User{ID: "1"}

	body := map[string]string{"password": "newpass"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/users/1/password", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, repo.data["1"].PasswordHash)
}

func TestUserService_Delete(t *testing.T) {
	app, repo := setupUserService()

	repo.data["1"] = &models.User{ID: "1"}

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
