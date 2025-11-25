package service

import (
	"github.com/gofiber/fiber/v2"
	"achievement_backend/app/repository"
)

type AchievementReferenceService struct {
	repo repository.AchievementReferenceRepository
}

func NewAchievementReferenceService(r repository.AchievementReferenceRepository) *AchievementReferenceService {
	return &AchievementReferenceService{repo: r}
}

//
// ====================== GET ALL ======================
//
func (s *AchievementReferenceService) GetAll(c *fiber.Ctx) error {
	data, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch references",
		})
	}
	return c.JSON(data)
}

//
// ====================== GET BY ID ======================
//
func (s *AchievementReferenceService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	data, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch reference",
		})
	}

	if data == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "reference not found",
		})
	}

	return c.JSON(data)
}

//
// ====================== GET BY STUDENT ======================
//
func (s *AchievementReferenceService) GetByStudent(c *fiber.Ctx) error {
	studentID := c.Params("student_id")

	data, err := s.repo.GetByStudentID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch student references",
		})
	}

	return c.JSON(data)
}

//
// ====================== CREATE REFERENCE ======================
//
// FR-003: Dibuat setelah Mongo draft dibuat
//
func (s *AchievementReferenceService) Create(c *fiber.Ctx) error {
	var req struct {
		StudentID string `json:"student_id"`
		MongoID   string `json:"mongo_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.StudentID == "" || req.MongoID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "student_id and mongo_id are required",
		})
	}

	data, err := s.repo.Create(req.StudentID, req.MongoID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to create achievement reference",
		})
	}

	return c.JSON(data)
}

//
// ====================== SUBMIT ======================
//
// FR-003: Mahasiswa hanya bisa submit jika status draft atau resubmitted
//
func (s *AchievementReferenceService) Submit(c *fiber.Ctx) error {
	id := c.Params("id")

	ref, err := s.repo.GetByID(id)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "reference not found",
		})
	}

	if ref.Status != "draft" && ref.Status != "resubmitted" {
		return c.Status(400).JSON(fiber.Map{
			"error": "status must be draft or resubmitted to submit",
		})
	}

	if err := s.repo.Submit(id); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to submit achievement",
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

//
// ====================== VERIFY ======================
//
// FR-004: Dosen pembimbing memverifikasi prestasi
//
func (s *AchievementReferenceService) Verify(c *fiber.Ctx) error {
	id := c.Params("id")
	verifierID := c.Locals("user_id").(string)

	ref, err := s.repo.GetByID(id)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "reference not found",
		})
	}

	if ref.Status != "submitted" {
		return c.Status(400).JSON(fiber.Map{
			"error": "only submitted achievements can be verified",
		})
	}

	if err := s.repo.Verify(id, verifierID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to verify achievement",
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

//
// ====================== REJECT ======================
//
// FR-004: Dosen pembimbing memberikan catatan penolakan
//
func (s *AchievementReferenceService) Reject(c *fiber.Ctx) error {
	id := c.Params("id")
	verifierID := c.Locals("user_id").(string)

	var req struct {
		Note string `json:"note"`
	}

	if err := c.BodyParser(&req); err != nil || req.Note == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "rejection note is required",
		})
	}

	ref, err := s.repo.GetByID(id)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "reference not found",
		})
	}

	if ref.Status != "submitted" {
		return c.Status(400).JSON(fiber.Map{
			"error": "only submitted achievements can be rejected",
		})
	}

	if err := s.repo.Reject(id, verifierID, req.Note); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to reject achievement",
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

//
// ====================== SOFT DELETE ======================
//
// FR-005: Hanya draft yang boleh dihapus (soft delete)
//
func (s *AchievementReferenceService) SoftDelete(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	ref, err := s.repo.GetByID(id)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "reference not found",
		})
	}

	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"error": "only draft achievements can be deleted",
		})
	}

	if err := s.repo.SoftDelete(id, userID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to delete achievement",
		})
	}

	return c.JSON(fiber.Map{"success": true})
}
