package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"time"

	models "achievement_backend/app/model"
	"achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type AchievementMongoService struct {
	mongoRepo    repository.MongoAchievementRepository
	refRepo      repository.AchievementReferenceRepository
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

// ==========================================
// CONSTRUCTOR
// ==========================================
func NewAchievementMongoService(
	mongo repository.MongoAchievementRepository,
	ref repository.AchievementReferenceRepository,
	student repository.StudentRepository,
	lecturer repository.LecturerRepository,
) *AchievementMongoService {
	return &AchievementMongoService{
		mongoRepo:    mongo,
		refRepo:      ref,
		studentRepo:  student,
		lecturerRepo: lecturer,
	}
}

// ===========================================================
//  1. LIST BY ROLE (Admin, Lecturer, Student) -- diperbaiki
//     Mengembalikan gabungan: metadata (Postgres reference) + detail (Mongo)
//     Mendukung pagination untuk admin/lecturer via query params ?page=&limit=
//
// ===========================================================
func (s *AchievementMongoService) ListByRole(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	roleName := c.Locals("role_name")
	if userID == nil || roleName == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized: missing session"})
	}
	uid := userID.(string)
	rid := roleName.(string)
	ctx := c.Context()

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	switch rid {
	case "Admin":
		// Admin: ambil references (postgres) dengan pagination, lalu ambil details batch dari mongo
		refs, total, err := s.refRepo.GetAllWithPagination(limit, offset)
		if err != nil {
			log.Printf("[ListByRole-Admin] refRepo.GetAllWithPagination error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch references"})
		}

		// build mongo ids
		mongoIDs := []string{}
		for _, r := range refs {
			if r.MongoAchievementID != "" {
				mongoIDs = append(mongoIDs, r.MongoAchievementID)
			}
		}

		mDetails, err := s.mongoRepo.GetManyByIDs(ctx, mongoIDs)
		if err != nil {
			log.Printf("[ListByRole-Admin] mongoRepo.GetManyByIDs error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievement details"})
		}

		// merge
		out := []fiber.Map{}
		for _, ref := range refs {
			detail, ok := mDetails[ref.MongoAchievementID]
			var det interface{}
			if ok {
				det = detail
			} else {
				det = nil
			}
			out = append(out, fiber.Map{
				"reference": ref,
				"detail":    det,
			})
		}

		return c.JSON(fiber.Map{
			"success":    true,
			"data":       out,
			"pagination": fiber.Map{"page": page, "limit": limit, "total": total},
		})

	case "Dosen Wali":
		// 1. Ambil data dosen berdasarkan user_id
		lecturer, err := s.lecturerRepo.GetByUserID(uid)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch lecturer"})
		}
		if lecturer == nil {
			return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
		}

		// 2. Ambil mahasiswa bimbingan berdasarkan lecturer.ID
		students, err := s.studentRepo.GetByAdvisorID(lecturer.ID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch advisees"})
		}

		// Jika tidak ada mahasiswa → hasil kosong
		if len(students) == 0 {
			return c.JSON(fiber.Map{
				"success": true,
				"data":    []interface{}{},
				"pagination": fiber.Map{
					"page":  page,
					"limit": limit,
					"total": 0,
				},
			})
		}

		// 3. Ambil list student.id untuk query reference
		studentIDs := make([]string, 0, len(students))
		for _, st := range students {
			studentIDs = append(studentIDs, st.ID)
		}

		// 4. Ambil achievement reference berdasarkan daftar mahasiswa
		refs, total, err := s.refRepo.GetByAdviseesWithPagination(studentIDs, limit, offset)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievements"})
		}

		// 5. Kumpulkan ID Mongo untuk batch fetch
		mongoIDs := make([]string, 0)
		for _, r := range refs {
			if r.MongoAchievementID != "" {
				mongoIDs = append(mongoIDs, r.MongoAchievementID)
			}
		}

		// 6. Ambil detail dari MongoDB sekaligus
		mDetails, err := s.mongoRepo.GetManyByIDs(ctx, mongoIDs)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievement details"})
		}

		// 7. Merge reference + detail
		out := make([]fiber.Map, 0, len(refs))
		for _, ref := range refs {
			detail, exists := mDetails[ref.MongoAchievementID]

			var det interface{}
			if exists {
				det = detail
			} else {
				det = nil
			}

			out = append(out, fiber.Map{
				"reference": ref,
				"detail":    det,
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data":    out,
			"pagination": fiber.Map{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		})

	case "Mahasiswa":
		// Mahasiswa: ambil student by user id, lalu ambil references and details for that student
		student, err := s.studentRepo.GetByUserID(uid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || student == nil {
				return c.Status(404).JSON(fiber.Map{"error": "student profile not found"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch student profile"})
		}

		refs, err := s.refRepo.GetByStudentID(student.ID)
		if err != nil {
			log.Printf("[ListByRole-Student] refRepo.GetByStudentID error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch student achievements"})
		}

		mongoIDs := []string{}
		for _, r := range refs {
			if r.MongoAchievementID != "" {
				mongoIDs = append(mongoIDs, r.MongoAchievementID)
			}
		}

		mDetails, err := s.mongoRepo.GetManyByIDs(ctx, mongoIDs)
		if err != nil {
			log.Printf("[ListByRole-Student] mongoRepo.GetManyByIDs error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievement details"})
		}

		out := []fiber.Map{}
		for _, ref := range refs {
			detail, ok := mDetails[ref.MongoAchievementID]
			var det interface{}
			if ok {
				det = detail
			} else {
				det = nil
			}
			out = append(out, fiber.Map{
				"reference": ref,
				"detail":    det,
			})
		}

		return c.JSON(fiber.Map{"success": true, "data": out})

	default:
		return c.Status(403).JSON(fiber.Map{"error": "unknown role / unauthorized role"})
	}
}

// ===========================================================
// 2. GET DETAIL -- gabungkan reference (postgres) + detail (mongo)
// ===========================================================
func (s *AchievementMongoService) GetDetail(c *fiber.Ctx) error {
	mongo_id := c.Params("id")
	ctx := c.Context()

	// Ambil detail dari Mongo
	item, err := s.mongoRepo.GetByID(ctx, mongo_id)
	if err != nil {
		log.Printf("[GetDetail] mongoRepo.GetByID error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievement detail"})
	}
	if item == nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}
	if item.IsDeleted {
		return c.Status(410).JSON(fiber.Map{"error": "achievement deleted"})
	}

	// Ambil reference dari Postgres (bila ada)
	ref, err := s.refRepo.GetByMongoAchievementID(mongo_id)
	if err != nil {
		// If not found, still return detail but warn
		log.Printf("[GetDetail] refRepo.GetByMongoAchievementID error: %v", err)
		return c.Status(200).JSON(fiber.Map{
			"success":   true,
			"reference": nil,
			"detail":    item,
			"warning":   "reference not found",
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"reference": ref,
		"detail":    item,
	})
}

// ===========================================================
// 3. CREATE DRAFT (FR-001) -- perbaikan: kompensasi jika ref gagal
// ===========================================================
func (s *AchievementMongoService) CreateDraft(c *fiber.Ctx) error {
	var req models.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	// Validate student exist
	// Validate student exist by ID (UUID students table)
	student, err := s.studentRepo.GetByID(req.StudentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(404).JSON(fiber.Map{"error": "student not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch student"})
	}
	if student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	ctx := c.Context()

	// Create in Mongo
	created, err := s.mongoRepo.CreateDraft(ctx, &req)
	if err != nil {
		log.Printf("[CreateDraft] mongoRepo.CreateDraft error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to create draft"})
	}

	mongoID := created.ID.Hex()

	// Create reference in Postgres - if fail, compensate by deleting Mongo doc
	ref, err := s.refRepo.Create(student.ID, mongoID)
	if err != nil {
		log.Printf("[CreateDraft] refRepo.Create error: %v", err)
		// compensate: soft-delete or delete the mongo document to avoid orphan
		_ = s.mongoRepo.SoftDelete(ctx, mongoID)
		return c.Status(500).JSON(fiber.Map{"error": "failed to create reference, draft rolled back"})
	}

	return c.Status(201).JSON(fiber.Map{
		"success":      true,
		"mongo_id":     mongoID,
		"reference_id": ref.ID,
		"detail":       created,
	})
}

// ===========================================================
// 4. UPDATE DRAFT (FR-002) -- perbaikan: ownership + single-update filter
// ===========================================================
func (s *AchievementMongoService) UpdateDraft(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	// ownership check: user must be the owner
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	uid := userID.(string)

	student, err := s.studentRepo.GetByUserID(uid)
	if err != nil || student == nil {
		return c.Status(403).JSON(fiber.Map{"error": "student profile not found or unauthorized"})
	}

	// Ensure the mongo document belongs to this student
	item, err := s.mongoRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Printf("[UpdateDraft] mongoRepo.GetByID error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievement"})
	}
	if item == nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}
	if item.StudentID != student.ID {
		return c.Status(403).JSON(fiber.Map{"error": "not the owner"})
	}

	// perform update (repo will enforce status == draft)
	updated, err := s.mongoRepo.UpdateDraft(c.Context(), id, &req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": updated})
}

// ===========================================================
// 5. DELETE DRAFT (FR-005) -- perbaikan: ownership + sync with reference
// ===========================================================
func (s *AchievementMongoService) DeleteDraft(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := c.Context()

	// ownership check
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	uid := userID.(string)
	student, err := s.studentRepo.GetByUserID(uid)
	if err != nil || student == nil {
		return c.Status(403).JSON(fiber.Map{"error": "student profile not found or unauthorized"})
	}

	// Ensure ownership
	item, err := s.mongoRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("[DeleteDraft] GetByID error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievement"})
	}
	if item == nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}
	if item.StudentID != student.ID {
		return c.Status(403).JSON(fiber.Map{"error": "not the owner"})
	}
	if item.Status != models.StatusDraft {
		return c.Status(400).JSON(fiber.Map{"error": "only draft achievements can be deleted"})
	}

	// First soft-delete in Mongo
	if err := s.mongoRepo.SoftDelete(ctx, id); err != nil {
		log.Printf("[DeleteDraft] mongoRepo.SoftDelete error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete achievement"})
	}

	// Then soft-delete reference in Postgres (if exists). If Postgres fails, try to rollback Mongo (best-effort).
	ref, err := s.refRepo.GetByMongoAchievementID(id)
	if err != nil {
		log.Printf("[DeleteDraft] refRepo.GetByMongoAchievementID error (continuing): %v", err)
		// return success because mongo is deleted; but warn client
		return c.JSON(fiber.Map{"success": true, "warning": "mongo deleted but reference deletion failed/found nil"})
	}
	if ref != nil {
		if err := s.refRepo.SoftDelete(ref.ID); err != nil {
			log.Printf("[DeleteDraft] refRepo.SoftDelete error: %v", err)
			// attempt rollback: unset mongo isDeleted (best-effort)
			_ = s.mongoRepo.UpdateStatus(ctx, id, models.StatusDraft)
			return c.Status(500).JSON(fiber.Map{"error": "failed to delete reference; rollback attempted"})
		}
	}

	return c.JSON(fiber.Map{"success": true, "message": "achievement marked as deleted"})
}

// ===========================================================
// 6. UPDATE ATTACHMENTS (FR-006) -- perbaikan: ownership + only draft
// ===========================================================
func (s *AchievementMongoService) UpdateAttachments(c *fiber.Ctx) error {
	id := c.Params("mongo_id")
	ctx := c.Context()

	// ownership check
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	uid := userID.(string)
	student, err := s.studentRepo.GetByUserID(uid)
	if err != nil || student == nil {
		return c.Status(403).JSON(fiber.Map{"error": "student profile not found or unauthorized"})
	}

	// Check existing document and ownership & status
	existing, err := s.mongoRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("[UpdateAttachments] GetByID error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievement"})
	}
	if existing == nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}
	if existing.StudentID != student.ID {
		return c.Status(403).JSON(fiber.Map{"error": "not the owner"})
	}
	if existing.Status != models.StatusDraft {
		return c.Status(400).JSON(fiber.Map{"error": "attachments can only be updated when draft"})
	}

	var attachments []models.Attachment
	contentType := c.Get("Content-Type")
	isMultipart := c.Is("multipart/form-data")

	if isMultipart || (contentType != "" && len(contentType) >= 19 && contentType[:19] == "multipart/form-data") {
		form, err := c.MultipartForm()
		if err != nil {
			log.Printf("Multipart parse error: %v", err)
			return c.Status(400).JSON(fiber.Map{"error": "invalid multipart form"})
		}

		var files []*multipart.FileHeader
		for _, fieldName := range []string{"attachments", "file", "files"} {
			if f, ok := form.File[fieldName]; ok && len(f) > 0 {
				files = f
				break
			}
		}
		if len(files) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "no files provided. Use field name: 'attachments', 'file', or 'files'"})
		}

		for _, fh := range files {
			dstName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fh.Filename))
			dst := filepath.Join("uploads", dstName)
			if err := c.SaveFile(fh, dst); err != nil {
				log.Printf("File save error: %v", err)
				return c.Status(500).JSON(fiber.Map{"error": "failed to save file: " + err.Error()})
			}
			attachments = append(attachments, models.Attachment{
				FileName:   fh.Filename,
				FileURL:    "/uploads/" + dstName,
				FileType:   fh.Header.Get("Content-Type"),
				UploadedAt: time.Now(),
			})
		}
	} else {
		var req models.UpdateAchievementAttachmentsRequest
		if err := c.BodyParser(&req); err != nil {
			log.Printf("Body parse error: %v", err)
			return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
		}
		if len(req.Attachments) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "attachments missing or null"})
		}
		attachments = req.Attachments
	}

	if len(attachments) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "no attachments to update"})
	}

	res, err := s.mongoRepo.UpdateAttachments(ctx, id, attachments)
	if err != nil {
		log.Printf("[UpdateAttachments] repo error: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "attachments updated", "data": res})
}

// ===========================================================
// 7. GET BY STUDENT (tidy wrapper)
// ===========================================================
func (s *AchievementMongoService) GetByStudent(c *fiber.Ctx) error {
	studentID := c.Params("id")
	ctx := c.Context()

	refs, err := s.refRepo.GetByStudentID(studentID)
	if err != nil {
		log.Printf("[GetByStudent] refRepo.GetByStudentID error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch student achievements"})
	}

	mongoIDs := []string{}
	for _, r := range refs {
		if r.MongoAchievementID != "" {
			mongoIDs = append(mongoIDs, r.MongoAchievementID)
		}
	}

	mDetails, err := s.mongoRepo.GetManyByIDs(ctx, mongoIDs)
	if err != nil {
		log.Printf("[GetByStudent] mongoRepo.GetManyByIDs error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievement details"})
	}

	out := []fiber.Map{}
	for _, ref := range refs {
		detail, ok := mDetails[ref.MongoAchievementID]
		var det interface{}
		if ok {
			det = detail
		} else {
			det = nil
		}
		out = append(out, fiber.Map{
			"reference": ref,
			"detail":    det,
		})
	}

	return c.JSON(fiber.Map{"success": true, "data": out})
}

// ===========================================================
// 8. Helper: expose update status wrapper (dipanggil dari ReferenceService)
// ===========================================================
func (s *AchievementMongoService) UpdateStatus(ctx context.Context, mongoID string, status string) error {
	return s.mongoRepo.UpdateStatus(ctx, mongoID, status)
}
