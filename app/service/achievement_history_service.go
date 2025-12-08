package service

import (
	"achievement_backend/app/repository"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AchievementHistoryService struct {
	refRepo   repository.AchievementReferenceRepository
	mongoRepo repository.MongoAchievementRepository
}

func NewAchievementHistoryService(
	refRepo repository.AchievementReferenceRepository,
	mongoRepo repository.MongoAchievementRepository,
) *AchievementHistoryService {
	return &AchievementHistoryService{
		refRepo:   refRepo,
		mongoRepo: mongoRepo,
	}
}

type HistoryEvent struct {
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
	VerifiedBy    *string   `json:"verified_by,omitempty"`
	RejectionNote *string   `json:"rejection_note,omitempty"`
	Description   string    `json:"description"`
}

// ================= GET HISTORY =================
func (s *AchievementHistoryService) GetHistory(c *fiber.Ctx) error {
	mongoAchievementID := c.Params("id")

	if mongoAchievementID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "achievement id is required",
		})
	}

	log.Printf("[HISTORY] Getting history for achievement: %s", mongoAchievementID)

	// Get achievement reference
	ref, err := s.refRepo.GetByMongoAchievementID(mongoAchievementID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to fetch achievement",
			"detail": err.Error(),
		})
	}

	// If not found
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement not found",
		})
	}

	// Build history
	history := []HistoryEvent{}

	// 1. Created (draft)
	history = append(history, HistoryEvent{
		Status:      "draft",
		Timestamp:   ref.CreatedAt,
		Description: "Achievement created as draft",
	})

	// 2. Submitted
	if ref.SubmittedAt != nil {
		history = append(history, HistoryEvent{
			Status:      "submitted",
			Timestamp:   *ref.SubmittedAt,
			Description: "Achievement submitted for verification",
		})
	}

	// 3. Verified or Rejected
	if ref.VerifiedAt != nil {
		switch ref.Status {
		case "verified":
			history = append(history, HistoryEvent{
				Status:      "verified",
				Timestamp:   *ref.VerifiedAt,
				VerifiedBy:  ref.VerifiedBy,
				Description: "Achievement verified by lecturer",
			})

		case "rejected":
			history = append(history, HistoryEvent{
				Status:        "rejected",
				Timestamp:     *ref.VerifiedAt,
				VerifiedBy:    ref.VerifiedBy,
				RejectionNote: ref.RejectionNote,
				Description:   "Achievement rejected by lecturer",
			})
		}
	}

	// 4. Deleted (soft delete)
	if ref.Status == "deleted" {
		history = append(history, HistoryEvent{
			Status:      "deleted",
			Timestamp:   ref.UpdatedAt,
			Description: "Achievement was deleted by student",
		})
	}

	log.Printf("[HISTORY] Retrieved %d history events", len(history))

	return c.JSON(fiber.Map{
		"success": true,
		"current": ref.Status,
		"data":    history,
	})
}

