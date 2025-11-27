package service

import (
	"achievement_backend/app/repository"
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
)

type AchievementReferenceService struct {
	repo repository.AchievementReferenceRepository
}

func NewAchievementReferenceService(r repository.AchievementReferenceRepository) *AchievementReferenceService {
	return &AchievementReferenceService{
		repo: r,
	}
}

// ================= GET ALL =================
func (s *AchievementReferenceService) GetAll(c *fiber.Ctx) error {
	data, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch references"})
	}
	return c.JSON(data)
}

// ================= GET BY ID =================
func (s *AchievementReferenceService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	data, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch reference"})
	}
	if data == nil {
		return c.Status(404).JSON(fiber.Map{"error": "reference not found"})
	}
	return c.JSON(data)
}

// ================= GET BY STUDENT =================
func (s *AchievementReferenceService) GetByStudent(c *fiber.Ctx) error {
	studentID := c.Params("student_id")
	data, err := s.repo.GetByStudentID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch student references"})
	}
	return c.JSON(data)
}

// ================= CREATE =================
func (s *AchievementReferenceService) Create(c *fiber.Ctx) error {
	var req struct {
		StudentID string `json:"student_id"`
		MongoID   string `json:"mongo_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.StudentID == "" || req.MongoID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "student_id and mongo_id are required"})
	}

	data, err := s.repo.Create(req.StudentID, req.MongoID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create achievement reference"})
	}
	return c.JSON(data)
}

// ================= SUBMIT =================
func (s *AchievementReferenceService) Submit(c *fiber.Ctx) error {
	mongoAchievementID := c.Params("id")

	if mongoAchievementID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "achievement id is required",
		})
	}

	log.Printf("[SUBMIT] Looking for achievement reference with mongo_id: %s", mongoAchievementID)

	// Cari achievement reference berdasarkan mongo achievement ID
	ref, err := s.repo.GetByMongoAchievementID(mongoAchievementID)
	if err != nil {
		log.Printf("[SUBMIT] Error fetching reference: %v", err)
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error":   "achievement reference not found",
				"mongoID": mongoAchievementID,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to fetch achievement reference",
			"detail": err.Error(),
		})
	}

	log.Printf("[SUBMIT] Found reference: %+v", ref)

	// Submit dengan reference ID yang benar
	err = s.repo.Submit(ref.ID)
	if err != nil {
		log.Printf("[SUBMIT] Error submitting: %v", err)
		if err == sql.ErrNoRows {
			return c.Status(400).JSON(fiber.Map{
				"error": "only draft achievements can be submitted",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to submit achievement",
			"detail": err.Error(),
		})
	}

	log.Printf("[SUBMIT] Successfully submitted reference: %s", ref.ID)
	return c.JSON(fiber.Map{"success": true})
}

// ================= VERIFY =================
func (s *AchievementReferenceService) Verify(c *fiber.Ctx) error {
	mongoAchievementID := c.Params("id")
	verifierID := c.Locals("user_id").(string)

	if mongoAchievementID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "achievement id is required",
		})
	}

	// Cari achievement reference berdasarkan mongo achievement ID
	ref, err := s.repo.GetByMongoAchievementID(mongoAchievementID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "achievement reference not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to fetch achievement reference",
			"detail": err.Error(),
		})
	}

	// Verify dengan reference ID yang benar
	err = s.repo.Verify(ref.ID, verifierID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(400).JSON(fiber.Map{"error": "only submitted achievements can be verified"})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to verify achievement",
			"detail": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true})
}

// ================= REJECT =================
func (s *AchievementReferenceService) Reject(c *fiber.Ctx) error {
	mongoAchievementID := c.Params("id")
	verifierID := c.Locals("user_id").(string)

	if mongoAchievementID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "achievement id is required",
		})
	}

	var req struct {
		Note string `json:"note"`
	}
	if err := c.BodyParser(&req); err != nil || req.Note == "" {
		return c.Status(400).JSON(fiber.Map{"error": "rejection note is required"})
	}

	// Cari achievement reference berdasarkan mongo achievement ID
	ref, err := s.repo.GetByMongoAchievementID(mongoAchievementID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "achievement reference not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to fetch achievement reference",
			"detail": err.Error(),
		})
	}

	// Reject dengan reference ID yang benar
	err = s.repo.Reject(ref.ID, verifierID, req.Note)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(400).JSON(fiber.Map{"error": "only submitted achievements can be rejected"})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to reject achievement",
			"detail": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true})
}

// ================= SOFT DELETE =================
func (s *AchievementReferenceService) SoftDelete(c *fiber.Ctx) error {
	mongoAchievementID := c.Params("id")

	if mongoAchievementID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "achievement id is required",
		})
	}

	// Cari achievement reference berdasarkan mongo achievement ID
	ref, err := s.repo.GetByMongoAchievementID(mongoAchievementID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "achievement reference not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to fetch achievement reference",
			"detail": err.Error(),
		})
	}

	// Delete dengan reference ID yang benar
	err = s.repo.SoftDelete(ref.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(400).JSON(fiber.Map{"error": "only draft achievements can be deleted"})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to delete achievement",
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"success": true})
}
