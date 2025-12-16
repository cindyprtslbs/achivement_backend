package service

import (
	"golang.org/x/crypto/bcrypt"

	models "achievement_backend/app/model"
	"achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type UserService struct {
	userRepo     repository.UserRepository
	roleRepo     repository.RoleRepository
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

// GetAllUsers godoc
// @Summary Mendapatkan daftar semua pengguna
// @Description Mendapatkan daftar semua pengguna
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Daftar pengguna"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil data"
// @Security Bearer
// @Router /api/v1/users [get]
func (s *UserService) GetAll(c *fiber.Ctx) error {
	users, err := s.userRepo.GetAll()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch users")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    users,
	})
}

// GetUserByID godoc
// @Summary Mendapatkan detail pengguna berdasarkan ID
// @Description Mendapatkan detail pengguna berdasarkan ID
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "Detail pengguna"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Security Bearer
// @Router /api/v1/users/{id} [get]
func (s *UserService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

// CreateUser godoc
// @Summary Membuat pengguna baru
// @Description Membuat pengguna baru
// @Tags User
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "Data pengguna baru"
// @Success 201 {object} map[string]interface{} "Pengguna berhasil dibuat"
// @Failure 400 {object} map[string]interface{} "Invalid request body atau data sudah ada"
// @Failure 500 {object} map[string]interface{} "Gagal membuat pengguna"
// @Security Bearer
// @Router /api/v1/users [post]
func (s *UserService) Create(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// validasi email
	if u, _ := s.userRepo.GetByEmail(req.Email); u != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Email already registered")
	}

	// validasi username
	if u, _ := s.userRepo.GetByUsername(req.Username); u != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Username already taken")
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to hash password")
	}
	req.PasswordHash = string(hashed)

	if req.RoleID != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "role_id cannot be set during user creation",
		})
	}

	user, err := s.userRepo.Create(req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create user")
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "User created",
		"data":    user,
	})
}

// UpdateUser godoc
// @Summary Memperbarui data pengguna
// @Description Memperbarui data pengguna berdasarkan ID
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body models.UpdateUserRequest true "Data pengguna yang diperbarui"
// @Success 200 {object} map[string]interface{} "Pengguna berhasil diperbarui"
// @Failure 400 {object} map[string]interface{} "Invalid request body atau data sudah ada"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Gagal memperbarui pengguna"
// @Security Bearer
// @Router /api/v1/users/{id} [put]
func (s *UserService) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	// cek user lama
	existing, err := s.userRepo.GetByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
	}

	if req.Username != nil {
		if u, _ := s.userRepo.GetByUsername(*req.Username); u != nil && u.ID != existing.ID {
			return fiber.NewError(fiber.StatusBadRequest, "Username already taken")
		}
		existing.Username = *req.Username
	}

	if req.Email != nil {
		if u, _ := s.userRepo.GetByEmail(*req.Email); u != nil && u.ID != existing.ID {
			return fiber.NewError(fiber.StatusBadRequest, "Email already registered")
		}
		existing.Email = *req.Email
	}

	if req.FullName != nil {
		existing.FullName = *req.FullName
	}

	if req.RoleID != nil {
		if _, err := s.roleRepo.GetByID(*req.RoleID); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid role_id")
		}
		existing.RoleID = req.RoleID
	}

	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	// ======================================
	// UPDATE USER DATA
	// ======================================
	updated, err := s.userRepo.UpdatePartial(existing)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update user")
	}

	// ======================================
	// UPDATE STUDENT PROFILE
	// ======================================
	if req.Student != nil {
		err := s.setStudentProfileFromUserUpdate(id, req.Student)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		// AMBIL DATA STUDENT TERBARU UNTUK RESPONSE
		student, _ := s.studentRepo.GetByUserID(id)

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Student profile updated",
			"data":    student,
		})
	}

	// ======================================
	// UPDATE LECTURER PROFILE
	// ======================================
	if req.Lecturer != nil {
		err := s.setLecturerProfileFromUserUpdate(id, req.Lecturer)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		lecturer, _ := s.lecturerRepo.GetByUserID(id)

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Lecturer profile updated",
			"data":    lecturer,
		})
	}

	// ======================================
	// DEFAULT RESPONSE → USER UPDATED
	// ======================================
	return c.JSON(fiber.Map{
		"success": true,
		"message": "User updated",
		"data":    updated,
	})
}

// UpdatePassword godoc
// @Summary Memperbarui password pengguna
// @Description Memperbarui password pengguna berdasarkan ID
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param password body map[string]string true "Password baru"
// @Success 200 {object} map[string]interface{} "Password berhasil diperbarui"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Gagal memperbarui password"
// @Security Bearer
// @Router /api/v1/users/{id}/password [put]
func (s *UserService) UpdatePassword(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid body")
	}

	if body.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Password is required")
	}

	// check user exists
	if _, err := s.userRepo.GetByID(id); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to hash password")
	}

	if err := s.userRepo.UpdatePassword(id, string(hashed)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update password")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Password updated",
	})
}

// DeleteUser godoc
// @Summary Menghapus user
// @Description Menghapus user berdasarkan ID
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "User berhasil dihapus"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Gagal menghapus user"
// @Security Bearer
// @Router /api/v1/users/{id} [delete]
func (s *UserService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if _, err := s.userRepo.GetByID(id); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	if err := s.userRepo.Delete(id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete user")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User deleted",
	})
}

// SetStudentProfile godoc
// @Summary Mengatur profil mahasiswa untuk pengguna
// @Description Mengatur profil mahasiswa untuk pengguna berdasarkan ID pengguna
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param profile body models.SetStudentProfileRequest true "Data profil mahasiswa"
// @Success 200 {object} map[string]interface{} "Profil mahasiswa berhasil diatur"
// @Failure 400 {object} map[string]interface{} "Invalid request body atau user bukan Mahasiswa"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Gagal mengatur profil mahasiswa"
// @Security Bearer
// @Router /api/v1/users/{id} [put]
func (s *UserService) setStudentProfileFromUserUpdate(userId string, data *models.SetStudentProfileRequest) error {
	if data == nil {
		return nil
	}

	user, err := s.userRepo.GetByID(userId)
	if err != nil {
		return err
	}

	if user.RoleName != "Mahasiswa" {
		return fiber.NewError(fiber.StatusBadRequest, "User is not Mahasiswa")
	}

	// check existing
	existing, err := s.studentRepo.GetByUserID(userId)
	if err != nil {
		return err
	}

	if existing != nil {
		_, err := s.studentRepo.Update(existing.ID, models.UpdateStudentRequest{
			UserID:       userId,
			StudentID:    data.StudentID,
			ProgramStudy: data.ProgramStudy,
			AcademicYear: data.AcademicYear,
			AdvisorID:    existing.AdvisorID,
		})
		return err
	}

	_, err = s.studentRepo.Create(models.CreateStudentRequest{
		UserID:       userId,
		StudentID:    data.StudentID,
		ProgramStudy: data.ProgramStudy,
		AcademicYear: data.AcademicYear,
		AdvisorID:    nil,
	})

	return err
}

// SetLecturerProfile godoc
// @Summary Mengatur profil dosen untuk pengguna
// @Description Mengatur profil dosen untuk pengguna berdasarkan ID pengguna
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param profile body models.SetLecturerProfileRequest true "Data profil dosen"
// @Success 200 {object} map[string]interface{} "Profil dosen berhasil diatur"
// @Failure 400 {object} map[string]interface{} "Invalid request body atau user bukan Dosen Wali"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Gagal mengatur profil dosen"
// @Security Bearer
// @Router /api/v1/users/{id} [put]
func (s *UserService) setLecturerProfileFromUserUpdate(userId string, data *models.SetLecturerProfileRequest) error {
	if data == nil {
		return nil
	}

	user, err := s.userRepo.GetByID(userId)
	if err != nil {
		return err
	}

	if user.RoleName != "Dosen Wali" {
		return fiber.NewError(fiber.StatusBadRequest, "User is not Dosen Wali")
	}

	existing, err := s.lecturerRepo.GetByUserID(userId)
	if err != nil {
		return err
	}

	if existing != nil {
		_, err := s.lecturerRepo.Update(existing.ID, models.UpdateLecturerRequest{
			LecturerID: data.LecturerID,
			Department: data.Department,
		})
		return err
	}

	_, err = s.lecturerRepo.Create(models.CreateLecturerRequest{
		UserID:     userId,
		LecturerID: data.LecturerID,
		Department: data.Department,
	})

	return err
}
