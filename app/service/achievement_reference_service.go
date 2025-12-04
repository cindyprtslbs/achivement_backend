package service

import (
	"achievement_backend/app/repository"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

type AchievementReferenceService struct {
	repo      repository.AchievementReferenceRepository
	mongoRepo repository.MongoAchievementRepository
}

func NewAchievementReferenceService(r repository.AchievementReferenceRepository, m repository.MongoAchievementRepository) *AchievementReferenceService {
	return &AchievementReferenceService{
		repo:      r,
		mongoRepo: m,
	}
}

// ================= GET ALL =================
func (s *AchievementReferenceService) GetAll(c *fiber.Ctx) error {

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Step 1 — Ambil metadata dari Postgres
	refs, total, err := s.repo.GetAllWithPagination(limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch references"})
	}

	// Step 2 — Ambil semua Mongo IDs
	mongoIDs := []string{}
	for _, r := range refs {
		mongoIDs = append(mongoIDs, r.MongoAchievementID)
	}

	// Step 3 — Fetch details dari MongoDB
	ctx := c.Context()
	mDetails, err := s.mongoRepo.GetManyByIDs(ctx, mongoIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch mongo data"})
	}

	// Step 4 — Merge
	result := []fiber.Map{}
	for _, ref := range refs {
		result = append(result, fiber.Map{
			"id":           ref.ID,
			"student_id":   ref.StudentID,
			"status":       ref.Status,
			"mongo_id":     ref.MongoAchievementID,
			"submitted_at": ref.SubmittedAt,
			"verified_at":  ref.VerifiedAt,
			"achievement":  mDetails[ref.MongoAchievementID],
		})
	}

	return c.JSON(fiber.Map{
		"data": result,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ================= GET BY ID =================
func (s *AchievementReferenceService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	ref, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch reference"})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "reference not found"})
	}

	mDetail, _ := s.mongoRepo.GetByID(c.Context(), ref.MongoAchievementID)

	return c.JSON(fiber.Map{
		"reference": ref,
		"details":   mDetail,
	})
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
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
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
    mongoID := c.Params("id") // ID dari route = MONGO ID

    // Cari reference melalui mongo_id
    ref, err := s.repo.GetByMongoAchievementID(mongoID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to get reference"})
    }
    if ref == nil {
        return c.Status(404).JSON(fiber.Map{"error": "reference not found"})
    }

    // Submit di Postgres
    if err := s.repo.Submit(ref.ID); err != nil {
        if err == sql.ErrNoRows {
            return c.Status(400).JSON(fiber.Map{"error": "only draft achievements can be submitted"})
        }
        return c.Status(500).JSON(fiber.Map{"error": "failed to submit"})
    }

    // Update status achievement di Mongo
    _ = s.mongoRepo.UpdateStatus(c.Context(), mongoID, "submitted")

    return c.JSON(fiber.Map{
		"success": true,
		"message": "Berhasil submit achievement",
	})
}

// ================= VERIFY =================
func (s *AchievementReferenceService) Verify(c *fiber.Ctx) error {
    mongoID := c.Params("id")
    verifierID := c.Locals("user_id").(string)

    // Cari reference berdasarkan MONGO ID
    ref, err := s.repo.GetByMongoAchievementID(mongoID)
    if err != nil || ref == nil {
        return c.Status(404).JSON(fiber.Map{"error": "reference not found"})
    }

    // Update reference ke verified
    if err := s.repo.Verify(ref.ID, verifierID); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "only submitted achievements can be verified"})
    }

    // Update status di MongoDB
    _ = s.mongoRepo.UpdateStatus(c.Context(), mongoID, "verified")

    return c.JSON(fiber.Map{
		"success": true,
		"message": "Berhasil verify achievement",
	})
}

// ================= REJECT =================
func (s *AchievementReferenceService) Reject(c *fiber.Ctx) error {
    mongoID := c.Params("id")
    verifierID := c.Locals("user_id").(string)

    var req struct {
        RejectionNote string `json:"rejection_note"`
    }
    if err := c.BodyParser(&req); err != nil || req.RejectionNote == "" {
        return c.Status(400).JSON(fiber.Map{"error": "rejection_note required"})
    }

    ref, err := s.repo.GetByMongoAchievementID(mongoID)
    if err != nil || ref == nil {
        return c.Status(404).JSON(fiber.Map{"error": "reference not found"})
    }

    // Reject reference
    if err := s.repo.Reject(ref.ID, verifierID, req.RejectionNote); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "only submitted achievements can be rejected"})
    }

    // Update Mongo
    _ = s.mongoRepo.UpdateStatus(c.Context(), mongoID, "rejected")

    return c.JSON(fiber.Map{
		"success": true,
		"message": "Berhasil reject achievement",
	})
}
