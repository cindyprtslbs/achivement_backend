package service

import (
	"achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type AchievementReferenceService struct {
	repo          repository.AchievementReferenceRepository
	mongoRepo     repository.MongoAchievementRepository
	studentRepo   repository.StudentRepository
	lecturerRepo  repository.LecturerRepository
}


func NewAchievementReferenceService(
	r repository.AchievementReferenceRepository,
	m repository.MongoAchievementRepository,
	s repository.StudentRepository,
	l repository.LecturerRepository,
) *AchievementReferenceService {
	return &AchievementReferenceService{
		repo:         r,
		mongoRepo:    m,
		studentRepo:  s,
		lecturerRepo: l,
	}
}

// GetAll godoc
// @Summary Get all achievement references
// @Description Mengambil daftar achievement reference dengan pagination dan detail dari MongoDB
// @Tags Achievement Reference
// @Accept json
// @Produce json
// @Param page query int false "Nomor halaman" default(1)
// @Param limit query int false "Jumlah data per halaman" default(10)
// @Success 200 {object} map[string]interface{} "Daftar achievement references"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil data"
// @Security Bearer
// @Router /api/v1/achievements [get]
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

// GetByID godoc
// @Summary Get achievement reference by ID
// @Description Mengambil satu achievement reference beserta detail MongoDB
// @Tags Achievement Reference
// @Accept json
// @Produce json
// @Param id path string true "Reference ID"
// @Success 200 {object} map[string]interface{} "Detail achievement reference"
// @Failure 404 {object} map[string]interface{} "Reference tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil data"
// @Security Bearer
// @Router /api/v1/achievements/{id} [get]
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

func (s *AchievementReferenceService) GetByStudent(c *fiber.Ctx) error {
	studentID := c.Params("student_id")

	data, err := s.repo.GetByStudentID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch student references"})
	}

	return c.JSON(data)
}

// Submit godoc
// @Summary Submit achievement
// @Description Mengirim achievement dari status draft ke submitted (Mahasiswa atau Admin)
// @Tags Achievement Reference
// @Accept json
// @Produce json
// @Param id path string true "Mongo Achievement ID"
// @Success 200 {object} map[string]interface{} "Achievement berhasil disubmit"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Reference tidak ditemukan"
// @Failure 400 {object} map[string]interface{} "Status tidak valid"
// @Failure 500 {object} map[string]interface{} "Gagal sinkronisasi MongoDB"
// @Security Bearer
// @Router /api/v1/achievements/{id}/submit [post]
func (s *AchievementReferenceService) Submit(c *fiber.Ctx) error {
	role := c.Locals("role_name").(string)
	mongoID := c.Params("id")

	ref, err := s.repo.GetByMongoAchievementID(mongoID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "reference not found"})
	}

	// ================= RBAC =================
	switch role {

	case "Admin":
		// full access

	case "Mahasiswa":
		userID := c.Locals("user_id").(string)
		student, err := s.studentRepo.GetByUserID(userID)
		if err != nil || student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "student not found"})
		}
		if ref.StudentID != student.ID {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
		}

	default:
		return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
	}

	// ================= UPDATE STATUS (ONCE) =================
	if err := s.repo.Submit(ref.ID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "only draft achievements can be submitted",
		})
	}

	if err := s.mongoRepo.UpdateStatus(
		c.Context(), mongoID, "submitted",
	); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to sync mongo status",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "achievement submitted",
		"data": fiber.Map{
			"id":           ref.ID,
			"status":       "submitted",
		},
	})
}

// Verify godoc
// @Summary Verify achievement
// @Description Memverifikasi achievement yang sudah disubmit (Admin atau Dosen Wali)
// @Tags Achievement Reference
// @Accept json
// @Produce json
// @Param id path string true "Mongo Achievement ID"
// @Success 200 {object} map[string]interface{} "Achievement berhasil diverifikasi"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Reference tidak ditemukan"
// @Failure 400 {object} map[string]interface{} "Status tidak valid"
// @Failure 500 {object} map[string]interface{} "Gagal sinkronisasi MongoDB"
// @Security Bearer
// @Router /api/v1/achievements/{id}/verify [post]
func (s *AchievementReferenceService) Verify(c *fiber.Ctx) error {
	role := c.Locals("role_name").(string)
	mongoID := c.Params("id")
	verifierID := c.Locals("user_id").(string)
	ctx := c.Context()

	ref, err := s.repo.GetByMongoAchievementID(mongoID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "reference not found"})
	}

	// RBAC
	switch role {
	case "Admin":
	case "Dosen Wali":
		lecturer, err := s.lecturerRepo.GetByUserID(verifierID)
		if err != nil || lecturer == nil {
			return c.Status(403).JSON(fiber.Map{"error": "lecturer not found"})
		}
		student, err := s.studentRepo.GetByID(ref.StudentID)
		if err != nil || student == nil {
			return c.Status(404).JSON(fiber.Map{"error": "student not found"})
		}
		if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
		}
	default:
		return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
	}

	// UPDATE POSTGRES
	if err := s.repo.Verify(ref.ID, verifierID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "only submitted achievements can be verified",
		})
	}

	// SYNC MONGO
	if err := s.mongoRepo.UpdateStatus(ctx, mongoID, "verified"); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to sync mongo status",
		})
	}

	// 🔥 RELOAD DATA (INI KUNCINYA)
	updatedRef, err := s.repo.GetByID(ref.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to reload reference",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "achievement verified",
		"data": fiber.Map{
			"id":          updatedRef.ID,
			"status":      updatedRef.Status,
			"verified_at": updatedRef.VerifiedAt,
			"verified_by": updatedRef.VerifiedBy,
		},
	})
}

// Reject godoc
// @Summary Reject achievement
// @Description Menolak achievement dengan catatan penolakan (Admin atau Dosen Wali)
// @Tags Achievement Reference
// @Accept json
// @Produce json
// @Param id path string true "Mongo Achievement ID"
// @Param body body object true "Rejection note" example({"rejection_note":"Data tidak valid"})
// @Success 200 {object} map[string]interface{} "Achievement berhasil ditolak"
// @Failure 400 {object} map[string]interface{} "Rejection note wajib diisi"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Reference tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Gagal sinkronisasi MongoDB"
// @Security Bearer
// @Router /api/v1/achievements/{id}/reject [post]
func (s *AchievementReferenceService) Reject(c *fiber.Ctx) error {
	role := c.Locals("role_name").(string)
	mongoID := c.Params("id")
	verifierID := c.Locals("user_id").(string)
	ctx := c.Context()

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

	// RBAC
	switch role {
	case "Admin":
	case "Dosen Wali":
		lecturer, err := s.lecturerRepo.GetByUserID(verifierID)
		if err != nil || lecturer == nil {
			return c.Status(403).JSON(fiber.Map{"error": "lecturer not found"})
		}
		student, err := s.studentRepo.GetByID(ref.StudentID)
		if err != nil || student == nil {
			return c.Status(404).JSON(fiber.Map{"error": "student not found"})
		}
		if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
		}
	default:
		return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
	}

	// UPDATE POSTGRES
	if err := s.repo.Reject(ref.ID, verifierID, req.RejectionNote); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "only submitted achievements can be rejected",
		})
	}

	if err := s.mongoRepo.UpdateStatus(ctx, mongoID, "rejected"); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to sync mongo status",
		})
	}

	updatedRef, err := s.repo.GetByID(ref.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to reload reference",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "achievement rejected",
		"data": fiber.Map{
			"id":             updatedRef.ID,
			"status":         updatedRef.Status,
			"rejection_note": updatedRef.RejectionNote,
		},
	})
}
