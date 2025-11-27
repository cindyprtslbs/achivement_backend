package service

import (
	"context"

	models "achievement_backend/app/model"
	"achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type AchievementMongoService struct {
	mongoRepo   repository.MongoAchievementRepository
	refRepo     repository.AchievementReferenceRepository
	studentRepo repository.StudentRepository
}

// ==========================================
// CONSTRUCTOR
// ==========================================
func NewAchievementMongoService(
	mongo repository.MongoAchievementRepository,
	ref repository.AchievementReferenceRepository,
	student repository.StudentRepository,
) *AchievementMongoService {
	return &AchievementMongoService{
		mongoRepo:   mongo,
		refRepo:     ref,
		studentRepo: student,
	}
}

// ===========================================================
// 1. LIST BY ROLE (Admin, Lecturer, Student)
// ===========================================================
func (s *AchievementMongoService) ListByRole(c *fiber.Ctx) error {

	// Ambil dari middleware — sesuai yang kamu set:
	userID := c.Locals("user_id")
	roleName := c.Locals("role_name")
	if userID == nil || roleName == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "unauthorized: missing user session",
		})
	}

	uid := userID.(string)
	rid := roleName.(string)

	ctx := context.Background()

	var list []models.Achievement
	var err error

	switch rid {
	case "Admin":
		list, err = s.mongoRepo.GetAll(ctx)

	case "Dosen Wali":
		students, _ := s.studentRepo.GetByAdvisorID(uid)
		var ids []string
		for _, st := range students {
			ids = append(ids, st.ID)
		}
		list, err = s.mongoRepo.GetByAdvisor(ctx, ids)

	case "Mahasiswa":
		student, _ := s.studentRepo.GetByUserID(uid)
		if student == nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "student profile not found",
			})
		}
		list, err = s.mongoRepo.GetByStudentID(ctx, student.ID)

	default:
		return c.Status(403).JSON(fiber.Map{
			"error": "unknown role / unauthorized role",
		})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievements"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    list,
	})
}

// ===========================================================
// 2. GET DETAIL
// ===========================================================
func (s *AchievementMongoService) GetDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := context.Background()

	item, err := s.mongoRepo.GetByID(ctx, id)
	if err != nil || item == nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}

	if item.IsDeleted {
		return c.Status(410).JSON(fiber.Map{"error": "achievement deleted"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    item,
	})
}

// ===========================================================
// 3. CREATE DRAFT (FR-001)
// ===========================================================
func (s *AchievementMongoService) CreateDraft(c *fiber.Ctx) error {
	var req models.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	// Validate student exist
	student, err := s.studentRepo.GetByID(req.StudentID)
	if err != nil || student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	ctx := context.Background()
	res, err := s.mongoRepo.CreateDraft(ctx, &req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create draft"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

// ===========================================================
// 4. UPDATE DRAFT (FR-002)
// ===========================================================
func (s *AchievementMongoService) UpdateDraft(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	ctx := context.Background()
	res, err := s.mongoRepo.UpdateDraft(ctx, id, &req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

// ===========================================================
// 5. DELETE DRAFT (FR-005)
// ===========================================================
func (s *AchievementMongoService) DeleteDraft(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := context.Background()

	if err := s.mongoRepo.SoftDelete(ctx, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete draft"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "achievement marked as deleted",
	})
}

// ===========================================================
// 6. UPDATE ATTACHMENTS (FR-006)
// ===========================================================
func (s *AchievementMongoService) UpdateAttachments(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateAchievementAttachmentsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	ctx := context.Background()
	res, err := s.mongoRepo.UpdateAttachments(ctx, id, req.Attachments)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "attachments updated",
		"data":    res,
	})
}

// ===========================================================
// 7. GET BY STUDENT
// ===========================================================
func (s *AchievementMongoService) GetByStudent(c *fiber.Ctx) error {
	studentID := c.Params("id")
	ctx := context.Background()

	list, err := s.mongoRepo.GetByStudentID(ctx, studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievements"})
	}

	return c.JSON(fiber.Map{"success": true, "data": list})
}
