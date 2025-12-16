package service

import (
	"achievement_backend/app/repository"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AchievementHistoryService struct {
	refRepo      repository.AchievementReferenceRepository
	mongoRepo    repository.MongoAchievementRepository
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

func NewAchievementHistoryService(
	refRepo repository.AchievementReferenceRepository,
	mongoRepo repository.MongoAchievementRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) *AchievementHistoryService {
	return &AchievementHistoryService{
		refRepo:      refRepo,
		mongoRepo:    mongoRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

type HistoryEvent struct {
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
	VerifiedBy    *string   `json:"verified_by,omitempty"`
	RejectionNote *string   `json:"rejection_note,omitempty"`
	Description   string    `json:"description"`
}

// GetHistory godoc
// @Summary Mendapatkan riwayat perubahan status prestasi
// @Description Mendapatkan riwayat perubahan status dari sebuah prestasi berdasarkan ID prestasi di MongoDB
// @Description Akses:
// @Description - Admin: semua data
// @Description - Mahasiswa: hanya achievement miliknya
// @Description - Dosen Wali: hanya achievement mahasiswa bimbingan
// @Tags Achievement History
// @Accept json
// @Produce json
// @Param id path string true "ID Prestasi di MongoDB"
// @Success 200 {object} map[string]interface{} "Riwayat perubahan status prestasi"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Failed to fetch achievement"
// @Security Bearer
// @Router /api/v1/achievements/{id}/history [get]
func (s *AchievementHistoryService) GetHistory(c *fiber.Ctx) error {
	mongoAchievementID := c.Params("id")

	role := c.Locals("role_name")
	userID := c.Locals("user_id")

	if role == nil || userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	r := role.(string)
	uid := userID.(string)

	// ================= GET REFERENCE =================
	ref, err := s.refRepo.GetByMongoAchievementID(mongoAchievementID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch achievement",
		})
	}

	if ref == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement not found",
		})
	}

	// ================= AUTHORIZATION =================
	switch r {

	case "Admin":
		// full access
		break

	case "Mahasiswa":
		student, err := s.studentRepo.GetByUserID(uid)
		if err != nil || student == nil {
			return c.Status(403).JSON(fiber.Map{
				"error": "student not found",
			})
		}

		if ref.StudentID != student.ID {
			return c.Status(403).JSON(fiber.Map{
				"error": "forbidden: not your achievement",
			})
		}

	case "Dosen Wali":
		lecturer, err := s.lecturerRepo.GetByUserID(uid)
		if err != nil || lecturer == nil {
			return c.Status(403).JSON(fiber.Map{
				"error": "lecturer not found",
			})
		}

		student, err := s.studentRepo.GetByID(ref.StudentID)
		if err != nil || student == nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "student not found",
			})
		}

		if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return c.Status(403).JSON(fiber.Map{
				"error": "forbidden: not your advisee",
			})
		}

	default:
		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden role",
		})
	}

	// ================= BUILD HISTORY =================
	history := []HistoryEvent{}

	history = append(history, HistoryEvent{
		Status:      "draft",
		Timestamp:   ref.CreatedAt,
		Description: "Achievement created as draft",
	})

	if ref.SubmittedAt != nil {
		history = append(history, HistoryEvent{
			Status:      "submitted",
			Timestamp:   *ref.SubmittedAt,
			Description: "Achievement submitted for verification",
		})
	}

	if ref.VerifiedAt != nil {
		if ref.Status == "verified" {
			history = append(history, HistoryEvent{
				Status:      "verified",
				Timestamp:   *ref.VerifiedAt,
				VerifiedBy:  ref.VerifiedBy,
				Description: "Achievement verified",
			})
		}

		if ref.Status == "rejected" {
			history = append(history, HistoryEvent{
				Status:        "rejected",
				Timestamp:     *ref.VerifiedAt,
				VerifiedBy:    ref.VerifiedBy,
				RejectionNote: ref.RejectionNote,
				Description:   "Achievement rejected",
			})
		}
	}

	if ref.Status == "deleted" {
		history = append(history, HistoryEvent{
			Status:      "deleted",
			Timestamp:   ref.UpdatedAt,
			Description: "Achievement was deleted",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"current": ref.Status,
		"data":    history,
	})
}
