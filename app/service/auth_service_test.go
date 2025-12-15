package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	models "achievement_backend/app/model"
	// "achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

//
// =======================================================
// MOCK REPOSITORIES (FULL INTERFACE)
// =======================================================
//

// ---------- USER ----------
type mockAuthUserRepo struct {
	users map[string]*models.User
}

func newMockAuthUserRepo() *mockAuthUserRepo {
	return &mockAuthUserRepo{users: map[string]*models.User{}}
}

func (m *mockAuthUserRepo) GetAll() ([]models.User, error) { return nil, nil }

func (m *mockAuthUserRepo) GetByID(id string) (*models.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, fiber.ErrNotFound
	}
	return u, nil
}

func (m *mockAuthUserRepo) GetByEmail(email string) (*models.User, error) { return nil, nil }

func (m *mockAuthUserRepo) GetByUsername(username string) (*models.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, fiber.ErrNotFound
}

func (m *mockAuthUserRepo) Create(req models.CreateUserRequest) (*models.User, error) {
	return nil, nil
}

func (m *mockAuthUserRepo) UpdatePartial(u *models.User) (*models.User, error) {
	return u, nil
}

func (m *mockAuthUserRepo) UpdatePassword(id string, hash string) error {
	return nil
}

func (m *mockAuthUserRepo) Delete(id string) error { return nil }

// ---------- ROLE ----------
type mockAuthRoleRepo struct{}

func (m *mockAuthRoleRepo) GetAll() ([]models.Role, error) { return nil, nil }

func (m *mockAuthRoleRepo) GetByID(id string) (*models.Role, error) {
	return &models.Role{ID: id, Name: "student"}, nil
}

func (m *mockAuthRoleRepo) Create(req models.CreateRoleRequest) (*models.Role, error) {
	return nil, nil
}

func (m *mockAuthRoleRepo) Update(id string, req models.UpdateRoleRequest) (*models.Role, error) {
	return nil, nil
}

func (m *mockAuthRoleRepo) Delete(id string) error { return nil }

// ---------- ROLE PERMISSION ----------
type mockRolePermRepo struct{}

func (m *mockRolePermRepo) AssignPermission(roleID string, permissionID string) error {
	return nil
}

func (m *mockRolePermRepo) RemovePermission(roleID string, permissionID string) error {
	return nil
}

func (m *mockRolePermRepo) GetPermissionsByRole(roleID string) ([]models.Permission, error) {
	return []models.Permission{
		{Name: "read"},
		{Name: "write"},
	}, nil
}

func (m *mockRolePermRepo) GetRolesByPermission(permissionID string) ([]models.Role, error) {
	return []models.Role{
		{ID: "1", Name: "student"},
	}, nil
}


// ---------- STUDENT ----------
type mockAuthStudentRepo struct{}

func (m *mockAuthStudentRepo) GetAll() ([]models.Student, error) { return nil, nil }
func (m *mockAuthStudentRepo) GetByID(id string) (*models.Student, error) {
	return nil, nil
}
func (m *mockAuthStudentRepo) GetByStudentID(studentID string) (*models.Student, error) {
	return nil, nil
}
func (m *mockAuthStudentRepo) GetByUserID(userID string) (*models.Student, error) {
	return &models.Student{UserID: userID}, nil
}
func (m *mockAuthStudentRepo) GetByAdvisorID(advisorID string) ([]models.Student, error) {
	return nil, nil
}
func (m *mockAuthStudentRepo) Create(req models.CreateStudentRequest) (*models.Student, error) {
	return nil, nil
}
func (m *mockAuthStudentRepo) Update(id string, req models.UpdateStudentRequest) (*models.Student, error) {
	return nil, nil
}
func (m *mockAuthStudentRepo) UpdateAdvisor(id string, advisorID string) error {
	return nil
}

// ---------- LECTURER ----------
type mockAuthLecturerRepo struct{}

func (m *mockAuthLecturerRepo) GetAll() ([]models.Lecturer, error) { return nil, nil }
func (m *mockAuthLecturerRepo) GetByID(id string) (*models.Lecturer, error) {
	return nil, nil
}
func (m *mockAuthLecturerRepo) GetByUserID(userID string) (*models.Lecturer, error) {
	return &models.Lecturer{UserID: userID}, nil
}
func (m *mockAuthLecturerRepo) GetByLecturerID(lecturerID string) (*models.Lecturer, error) {
	return nil, nil
}
func (m *mockAuthLecturerRepo) Create(req models.CreateLecturerRequest) (*models.Lecturer, error) {
	return nil, nil
}
func (m *mockAuthLecturerRepo) Update(id string, req models.UpdateLecturerRequest) (*models.Lecturer, error) {
	return nil, nil
}

//
// =======================================================
// SETUP
// =======================================================
//

func setupAuthService() (*fiber.App, *mockAuthUserRepo) {
	app := fiber.New()

	userRepo := newMockAuthUserRepo()

	service := NewAuthService(
		userRepo,
		&mockAuthRoleRepo{},
		&mockRolePermRepo{},
		&mockAuthStudentRepo{},
		&mockAuthLecturerRepo{},
	)

	app.Post("/login", service.Login)
	app.Post("/refresh", service.RefreshToken)
	app.Post("/logout", service.Logout)
	app.Get("/profile", func(c *fiber.Ctx) error {
		c.Locals("user_id", "1")
		c.Locals("role_name", "student")
		return service.GetProfile(c)
	})

	return app, userRepo
}

//
// =======================================================
// TEST LOGIN SUCCESS
// =======================================================
//

func TestAuthService_Login_Success(t *testing.T) {
	app, repo := setupAuthService()

	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)

	repo.users["1"] = &models.User{
		ID:           "1",
		Username:     "john",
		PasswordHash: string(hash),
		IsActive:     true,
		RoleID:       ptr("role-1"),
		FullName:     "John Doe",
	}

	body := models.LoginRequest{
		Username: "john",
		Password: "secret",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

//
// =======================================================
// TEST REFRESH TOKEN
// =======================================================
//

func TestAuthService_RefreshToken(t *testing.T) {
	app, repo := setupAuthService()

	repo.users["1"] = &models.User{
		ID:       "1",
		RoleID:   ptr("role-1"),
		IsActive: true,
	}

	refreshMutex.Lock()
	refreshTokens["valid"] = refreshEntry{
		UserID:    "1",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	refreshMutex.Unlock()

	body := map[string]string{"refresh_token": "valid"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

//
// =======================================================
// TEST GET PROFILE
// =======================================================
//

func TestAuthService_GetProfile(t *testing.T) {
	app, repo := setupAuthService()

	repo.users["1"] = &models.User{
		ID:       "1",
		Username: "john",
		FullName: "John Doe",
		IsActive: true,
	}

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

//
// =======================================================
// HELPER
// =======================================================
//

func ptr[T any](v T) *T {
	return &v
}
