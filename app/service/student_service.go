package service

import (
	models "achievement_backend/app/model"
	"achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type StudentService struct {
	repo repository.StudentRepository
}

func NewStudentService(r repository.StudentRepository) *StudentService {
	return &StudentService{repo: r}
}

// ==========================================
// GET ALL STUDENTS (role-aware)
// ==========================================
func (s *StudentService) GetAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	roleName := c.Locals("role_name")

	if userID == nil || roleName == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized: missing user session"})
	}

	uid := userID.(string)
	role := roleName.(string)

	var students []models.Student
	var err error

	switch role {
	case "Admin":
		// Admin bisa lihat semua mahasiswa
		students, err = s.repo.GetAll()
	case "Dosen":
		// Dosen hanya lihat mahasiswa yang dibimbing
		students, err = s.repo.GetByAdvisorID(uid)
	case "Mahasiswa":
		// Mahasiswa hanya lihat dirinya sendiri
		student, _ := s.repo.GetByUserID(uid)
		if student != nil {
			students = []models.Student{*student}
		} else {
			students = []models.Student{}
		}
	default:
		return c.Status(403).JSON(fiber.Map{"error": "unauthorized role"})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch students"})
	}

	return c.JSON(fiber.Map{"success": true, "data": students})
}

// ==========================================
// GET STUDENT BY ID
// ==========================================
func (s *StudentService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	student, err := s.repo.GetByID(id)
	if err != nil || student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	return c.JSON(fiber.Map{"success": true, "data": student})
}

// ==========================================
// GET STUDENT BY USER ID
// ==========================================
func (s *StudentService) GetByUserID(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	student, err := s.repo.GetByUserID(userID)
	if err != nil || student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	return c.JSON(fiber.Map{"success": true, "data": student})
}

// ==========================================
// GET STUDENTS BY ADVISOR (Lecturer)
// ==========================================
func (s *StudentService) GetByAdvisorID(c *fiber.Ctx) error {
	advisorID := c.Params("advisor_id")

	data, err := s.repo.GetByAdvisorID(advisorID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get students"})
	}

	return c.JSON(fiber.Map{"success": true, "data": data})
}

// ==========================================
// CREATE STUDENT (Admin)
// ==========================================
func (s *StudentService) Create(c *fiber.Ctx) error {
	var req models.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	student, err := s.repo.Create(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create student"})
	}

	return c.Status(201).JSON(fiber.Map{"success": true, "message": "student created", "data": student})
}

// ==========================================
// UPDATE STUDENT (Admin)
// ==========================================
func (s *StudentService) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	student, err := s.repo.Update(id, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update student"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "student updated", "data": student})
}

// ==========================================
// DELETE STUDENT (Admin)
// ==========================================
func (s *StudentService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.repo.Delete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete student"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "student deleted"})
}

// ==========================================
// UPDATE ADVISOR (Admin / System)
// ==========================================
func (s *StudentService) UpdateAdvisor(c *fiber.Ctx) error {
	id := c.Params("id")

	type AdvisorRequest struct {
		AdvisorID string `json:"advisor_id"`
	}

	var req AdvisorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	// cek apakah student ada
	student, err := s.repo.GetByID(id)
	if err != nil || student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	updateReq := models.UpdateStudentRequest{
		UserID:       student.UserID,
		StudentID:    student.StudentID,
		ProgramStudy: student.ProgramStudy,
		AcademicYear: student.AcademicYear,
		AdvisorID:    req.AdvisorID,
	}

	updated, err := s.repo.Update(id, updateReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update advisor"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "advisor updated", "data": updated})
}
