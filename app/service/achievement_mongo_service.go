package service

import (
	"context"

	models "achievement_backend/app/model"
	"achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type AchievementMongoService struct {
	repo repository.MongoAchievementRepository
}

func NewAchievementMongoService(r repository.MongoAchievementRepository) *AchievementMongoService {
	return &AchievementMongoService{repo: r}
}

// ================= CREATE DRAFT =================
// FR-001: Create draft prestasi
func (s *AchievementMongoService) CreateDraft(c *fiber.Ctx) error {
	var req models.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	result, err := s.repo.CreateDraft(context.Background(), &req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create draft"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// ================= GET BY ID =================
func (s *AchievementMongoService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	result, err := s.repo.GetByID(context.Background(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get achievement"})
	}

	if result == nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}

	return c.JSON(result)
}

// ================= GET BY STUDENT ID =================
func (s *AchievementMongoService) GetByStudent(c *fiber.Ctx) error {
	studentID := c.Params("student_id")

	result, err := s.repo.GetByStudentID(context.Background(), studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievements"})
	}

	return c.JSON(result)
}

// ================= UPDATE DRAFT =================
// FR-002: Update draft, hanya jika status draft
func (s *AchievementMongoService) UpdateDraft(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	result, err := s.repo.UpdateDraft(context.Background(), id, &req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// ================= UPDATE ATTACHMENTS =================
// FR-006: Update file hanya ketika draft
func (s *AchievementMongoService) UpdateAttachments(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Attachments []models.Attachment `json:"attachments"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	result, err := s.repo.UpdateAttachments(context.Background(), id, req.Attachments)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// ================= SOFT DELETE =================
// FR-005: Hapus prestasi draft (soft delete)
func (s *AchievementMongoService) SoftDelete(c *fiber.Ctx) error {
	id := c.Params("id")

	err := s.repo.SoftDelete(context.Background(), id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}
