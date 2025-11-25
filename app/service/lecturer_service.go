package service

import (
	models "achievement_backend/app/model"
	"achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type LecturerService struct {
	repo repository.LecturerRepository
}

func NewLecturerService(r repository.LecturerRepository) *LecturerService {
	return &LecturerService{repo: r}
}


// GET ALL LECTURERS

func (s *LecturerService) GetAll(c *fiber.Ctx) error {
	data, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get lecturers"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}


// GET LECTURER BY ID

func (s *LecturerService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	lecturer, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    lecturer,
	})
}


// GET LECTURER BY USER_ID

func (s *LecturerService) GetByUserID(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	lecturer, err := s.repo.GetByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    lecturer,
	})
}


// CREATE LECTURER (ADMIN)

func (s *LecturerService) Create(c *fiber.Ctx) error {
	var req models.CreateLecturerRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	lecturer, err := s.repo.Create(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create lecturer"})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "lecturer created",
		"data":    lecturer,
	})
}


// UPDATE LECTURER (ADMIN)

func (s *LecturerService) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateLecturerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	lecturer, err := s.repo.Update(id, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update lecturer"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "lecturer updated",
		"data":    lecturer,
	})
}


// DELETE LECTURER (ADMIN)

func (s *LecturerService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.repo.Delete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete lecturer"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "lecturer deleted",
	})
}
