package service

import (
	// models "achievement_backend/app/model"
	"achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type StudentService struct {
	repo         repository.StudentRepository
	userRepo     repository.UserRepository
	lecturerRepo repository.LecturerRepository
}

func NewStudentService(
	s repository.StudentRepository,
	u repository.UserRepository,
	l repository.LecturerRepository,
) *StudentService {
	return &StudentService{
		repo:         s,
		userRepo:     u,
		lecturerRepo: l,
	}
}

// GetAllStudents godoc
// @Summary Mendapatkan daftar semua mahasiswa
// @Description Mendapatkan daftar semua mahasiswa (hanya untuk Admin)
// @Tags Student
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Daftar mahasiswa"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil data"
// @Security Bearer
// @Router /api/v1/students [get]
func (s *StudentService) GetAll(c *fiber.Ctx) error {
	roleName := c.Locals("role_name")
	if roleName == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	role := roleName.(string)

	// ================= ONLY ADMIN =================
	if role != "Admin" {
		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden: only admin can access this resource",
		})
	}

	students, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to get students",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    students,
	})
}

// GetStudentByID godoc
// @Summary Mendapatkan detail mahasiswa berdasarkan ID
// @Description Mendapatkan detail mahasiswa berdasarkan ID (hanya untuk Admin)
// @Tags Student
// @Accept json
// @Produce json
// @Param id path string true "ID Mahasiswa"
// @Success 200 {object} map[string]interface{} "Detail mahasiswa"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Mahasiswa tidak ditemukan"
// @Security Bearer
// @Router /api/v1/students/{id} [get]
func (s *StudentService) GetByID(c *fiber.Ctx) error {
	roleName := c.Locals("role_name")
	if roleName == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	role := roleName.(string)

	// ================= ONLY ADMIN =================
	if role != "Admin" {
		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden: only admin can access student detail",
		})
	}

	id := c.Params("id")

	student, err := s.repo.GetByID(id)
	if err != nil || student == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "student not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    student,
	})
}

// UpdateAdvisor godoc
// @Summary Memperbarui dosen wali mahasiswa
// @Description Memperbarui dosen wali dari seorang mahasiswa (hanya untuk Admin)
// @Tags Student
// @Accept json
// @Produce json
// @Param id path string true "ID Mahasiswa"
// @Param advisor body map[string]string true "ID Dosen Wali Baru" example({"advisor_id": "lecturer123"})
// @Success 200 {object} map[string]interface{} "Dosen wali berhasil diperbarui"
// @Failure 400 {object} map[string]interface{} "advisor_id is required"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Mahasiswa tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Gagal memperbarui dosen wali"
// @Security Bearer
// @Router /api/v1/students/{id}/advisor [put]
func (s *StudentService) UpdateAdvisor(c *fiber.Ctx) error {
	roleName := c.Locals("role_name")
	if roleName == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	role := roleName.(string)

	// ================= ONLY ADMIN =================
	if role != "Admin" {
		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden: only admin can update advisor",
		})
	}

	id := c.Params("id")

	type AdvisorRequest struct {
		AdvisorID string `json:"advisor_id"`
	}

	var req AdvisorRequest
	if err := c.BodyParser(&req); err != nil || req.AdvisorID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "advisor_id is required",
		})
	}

	student, err := s.repo.GetByID(id)
	if err != nil || student == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "student not found",
		})
	}

	// Update only advisor_id
	if err := s.repo.UpdateAdvisor(id, req.AdvisorID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to update advisor",
		})
	}

	updated, _ := s.repo.GetByID(id)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "advisor updated successfully",
		"data":    updated,
	})
}
