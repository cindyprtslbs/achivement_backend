package service

import (
	"achievement_backend/app/repository"
	"database/sql"
	"log"

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
    refID := c.Params("ref_id")

    // get reference by ref_id
    ref, err := s.repo.GetByID(refID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to get reference"})
    }
    if ref == nil {
        return c.Status(404).JSON(fiber.Map{"error": "reference not found"})
    }

    // submit in postgres
    if err := s.repo.Submit(ref.ID); err != nil {
        if err == sql.ErrNoRows {
            return c.Status(400).JSON(fiber.Map{"error": "only draft achievements can be submitted"})
        }
        return c.Status(500).JSON(fiber.Map{"error": "failed to submit"})
    }

    // update mongo
    mongoID := ref.MongoAchievementID
    _ = s.mongoRepo.UpdateStatus(c.Context(), mongoID, "submitted")

    return c.JSON(fiber.Map{"success": true})
}

// ================= VERIFY =================
func (s *AchievementReferenceService) Verify(c *fiber.Ctx) error {
    // AMBIL ref_id dari route
    refID := c.Params("ref_id")
    if refID == "" {
        return c.Status(400).JSON(fiber.Map{"error": "reference id is required"})
    }

    verifierID := c.Locals("user_id").(string)

    // Ambil reference berdasarkan ref_id (Postgres)
    ref, err := s.repo.GetByID(refID)
    if err != nil {
        if err == sql.ErrNoRows {
            return c.Status(404).JSON(fiber.Map{"error": "reference not found"})
        }
        return c.Status(500).JSON(fiber.Map{"error": "failed to fetch reference"})
    }

    // Update status di PostgreSQL
    if err := s.repo.Verify(refID, verifierID); err != nil {
        if err == sql.ErrNoRows {
            return c.Status(400).JSON(fiber.Map{"error": "only submitted achievements can be verified"})
        }
        return c.Status(500).JSON(fiber.Map{"error": "failed to verify reference"})
    }

    // Update status di Mongo
    if err := s.mongoRepo.UpdateStatus(c.Context(), ref.MongoAchievementID, "verified"); err != nil {
        log.Printf("[WARN] Failed updating mongo achievement: %v", err)
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": "achievement verified",
    })
}

// ================= REJECT =================
func (s *AchievementReferenceService) Reject(c *fiber.Ctx) error {
    refID := c.Params("ref_id")
    if refID == "" {
        return c.Status(400).JSON(fiber.Map{"error": "reference id is required"})
    }

    verifierID := c.Locals("user_id").(string)

    var req struct {
        RejectionNote string `json:"rejection_note"`
    }
    if err := c.BodyParser(&req); err != nil || req.RejectionNote == "" {
        return c.Status(400).JSON(fiber.Map{"error": "rejection_note required"})
    }

    ref, err := s.repo.GetByID(refID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to get reference"})
    }
    if ref == nil {
        return c.Status(404).JSON(fiber.Map{"error": "reference not found"})
    }

    // Update Postgres
    if err := s.repo.Reject(ref.ID, verifierID, req.RejectionNote); err != nil {
        if err == sql.ErrNoRows {
            return c.Status(400).JSON(fiber.Map{"error": "only submitted achievements can be rejected"})
        }
        return c.Status(500).JSON(fiber.Map{"error": "failed to reject"})
    }

    // Update MongoDB (FIXED)
    if err := s.mongoRepo.UpdateStatus(c.Context(), ref.MongoAchievementID, "rejected"); err != nil {
        log.Printf("[WARN] Mongo update failed: %v", err)
    }

    return c.JSON(fiber.Map{"success": true})
}

// ================= DELETE =================
func (s *AchievementReferenceService) SoftDelete(c *fiber.Ctx) error {
	mongoID := c.Params("id")

	ref, err := s.repo.GetByMongoAchievementID(mongoID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get reference"})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "reference not found"})
	}

	if err := s.repo.SoftDelete(ref.ID); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(400).JSON(fiber.Map{"error": "only draft achievements can be deleted"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete"})
	}

	// Update MongoDB
	if err := s.mongoRepo.UpdateStatus(c.Context(), mongoID, "deleted"); err != nil {
		log.Printf("[WARN] Mongo update failed: %v", err)
	}

	return c.JSON(fiber.Map{"success": true})
}
