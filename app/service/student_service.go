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


// GET ALL STUDENTS

func (s *StudentService) GetAll(c *fiber.Ctx) error {
	data, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get students"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}


// GET STUDENT BY ID

func (s *StudentService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	student, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    student,
	})
}


// GET STUDENT BY USER ID

func (s *StudentService) GetByUserID(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	student, err := s.repo.GetByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    student,
	})
}


// GET STUDENTS BY ADVISOR (Lecturer)

func (s *StudentService) GetByAdvisorID(c *fiber.Ctx) error {
	advisorID := c.Params("advisor_id")

	data, err := s.repo.GetByAdvisorID(advisorID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get students"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}


// CREATE STUDENT (Admin)

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


// UPDATE STUDENT (Admin)

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

	return c.JSON(fiber.Map{
		"success": true,
		"message": "student updated",
		"data":    student,
	})
}


// DELETE STUDENT (Admin)

func (s *StudentService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.repo.Delete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete student"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "student deleted",
	})
}
