package service

import (
	models "achievement_backend/app/model"
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

// ======================================================
// GET ALL STUDENTS (Role Aware)
// ======================================================
func (s *StudentService) GetAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	roleName := c.Locals("role_name")

	if userID == nil || roleName == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized: missing session"})
	}

	uid := userID.(string)
	role := roleName.(string)

	var (
		students []models.Student
		err      error
	)

	switch role {
	case "Admin":
		students, err = s.repo.GetAll()

	case "Dosen Wali":
		lecturer, err := s.lecturerRepo.GetByUserID(uid)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to get lecturer"})
		}
		if lecturer == nil {
			return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
		}

		// students, err = s.repo.GetByAdvisorID(lecturer.ID)

	case "Mahasiswa":
		student, _ := s.repo.GetByUserID(uid)
		if student != nil {
			students = []models.Student{*student}
		} else {
			students = []models.Student{}
		}

	default:
		return c.Status(403).JSON(fiber.Map{"error": "forbidden role"})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get students"})
	}

	return c.JSON(fiber.Map{"success": true, "data": students})
}

// ======================================================
// GET STUDENT BY ID
// ======================================================
func (s *StudentService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	student, err := s.repo.GetByID(id)
	if err != nil || student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	return c.JSON(fiber.Map{"success": true, "data": student})
}

// ======================================================
// GET STUDENT BY USER ID
// ======================================================
func (s *StudentService) GetByUserID(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	student, err := s.repo.GetByUserID(userID)
	if err != nil || student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	return c.JSON(fiber.Map{"success": true, "data": student})
}

// ======================================================
// GET STUDENTS BY ADVISOR
// ======================================================
func (s *StudentService) GetByAdvisorID(c *fiber.Ctx) error {
	advisorID := c.Params("advisor_id")

	data, err := s.repo.GetByAdvisorID(advisorID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get students"})
	}

	return c.JSON(fiber.Map{"success": true, "data": data})
}

// ======================================================
// CREATE STUDENT (Admin)
// ======================================================
func (s *StudentService) Create(c *fiber.Ctx) error {
	var req models.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	student, err := s.repo.Create(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create student"})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "student created",
		"data":    student,
	})
}

// ======================================================
// UPDATE STUDENT (Admin)
// ======================================================
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

	return c.JSON(fiber.Map{"success": true, "message": "updated", "data": student})
}

// ======================================================
// DELETE STUDENT
// ======================================================
func (s *StudentService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.repo.Delete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete student"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "student deleted"})
}

// ======================================================
// UPDATE ADVISOR ONLY
// ======================================================
func (s *StudentService) UpdateAdvisor(c *fiber.Ctx) error {
	id := c.Params("id")

	type AdvisorRequest struct {
		AdvisorID string `json:"advisor_id"`
	}

	var req AdvisorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	student, err := s.repo.GetByID(id)
	if err != nil || student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	// update only advisor_id
	err = s.repo.UpdateAdvisor(id, req.AdvisorID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update advisor"})
	}

	updated, _ := s.repo.GetByID(id)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "advisor updated successfully",
		"data":    updated,
	})
}

// ======================================================
// SET STUDENT PROFILE (Mahasiswa Only, No Advisor)
// ======================================================
func (s *StudentService) SetStudentProfile(c *fiber.Ctx) error {
	userId := c.Params("id")

	var req models.SetStudentProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Check user exists
	user, err := s.userRepo.GetByID(userId)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	// Must be Mahasiswa
	if user.RoleName != "Mahasiswa" {
		return fiber.NewError(fiber.StatusBadRequest, "User is not Mahasiswa")
	}

	// Check existing profile
	existing, err := s.repo.GetByUserID(userId)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// UPDATE profile
	if existing != nil {
		updated, err := s.repo.Update(existing.ID, models.UpdateStudentRequest{
			UserID:       userId,
			StudentID:    req.StudentID,
			ProgramStudy: req.ProgramStudy,
			AcademicYear: req.AcademicYear,
			AdvisorID:    existing.AdvisorID, // keep old advisor
		})
		if err != nil {
			return fiber.ErrInternalServerError
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Student profile updated successfully",
			"data":    updated,
		})
	}

	// CREATE new profile
	created, err := s.repo.Create(models.CreateStudentRequest{
		UserID:       userId,
		StudentID:    req.StudentID,
		ProgramStudy: req.ProgramStudy,
		AcademicYear: req.AcademicYear,
		AdvisorID:    nil, // advisor belum di-set
	})
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Student profile created successfully",
		"data":    created,
	})
}
