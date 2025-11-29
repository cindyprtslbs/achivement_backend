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
	student, err := s.studentRepo.GetByStudentID(req.StudentID)
	if err != nil {
		// If not found by student_id (e.g. the client passed the UUID `id`),
		// try lookup by ID. Repository returns sql.ErrNoRows when not found.
		if errors.Is(err, sql.ErrNoRows) {
			student, err = s.studentRepo.GetByID(req.StudentID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return c.Status(404).JSON(fiber.Map{"error": "student not found"})
				}
				return c.Status(500).JSON(fiber.Map{"error": "failed to fetch student"})
			}
		} else {
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch student"})
		}
	}

	if student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	ctx := context.Background()
	res, err := s.mongoRepo.CreateDraft(ctx, &req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create draft"})
	}

	// Create achievement reference di PostgreSQL
	mongoIDStr := res.ID.Hex()
	log.Printf("Created achievement in MongoDB with ID: %s", mongoIDStr)
	log.Printf("Student UUID: %s", student.ID)

	_, err = s.refRepo.Create(student.ID, mongoIDStr)
	if err != nil {
		log.Printf("Error creating reference: %v", err)
		// Log error tapi jangan buat response error - achievement sudah dibuat di MongoDB
		// Hanya reference yang gagal, nanti bisa retry
		return c.Status(201).JSON(fiber.Map{
			"success": true,
			"data":    res,
			"warning": "achievement created but reference creation failed: " + err.Error(),
		})
	}

	log.Printf("[CREATE] Successfully created reference for achievement %s", mongoIDStr)
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

	ctx := context.Background()

	var attachments []models.Attachment

	// Support multipart file upload (form field `attachments` as files)
	contentType := c.Get("Content-Type")
	isMultipart := c.Is("multipart/form-data")

	if isMultipart || (contentType != "" && contentType[:19] == "multipart/form-data") {
		form, err := c.MultipartForm()
		if err != nil {
			log.Printf("Multipart parse error: %v", err)
			return c.Status(400).JSON(fiber.Map{"error": "invalid multipart form"})
		}

		// Try multiple field names (attachments, file, files)
		var files []*multipart.FileHeader
		for _, fieldName := range []string{"attachments", "file", "files"} {
			if f, ok := form.File[fieldName]; ok && len(f) > 0 {
				files = f
				log.Printf("Found files under field name: %s (count: %d)", fieldName, len(f))
				break
			}
		}

		if len(files) == 0 {
			// Log all available fields for debugging
			availableFields := []string{}
			for key := range form.File {
				availableFields = append(availableFields, key)
			}
			log.Printf("No files found. Available file fields: %v", availableFields)
			return c.Status(400).JSON(fiber.Map{
				"error": "no files provided. Use field name: 'attachments', 'file', or 'files'",
			})
		}

		for _, fh := range files {
			// generate destination path and save file
			dstName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fh.Filename))
			dst := filepath.Join("uploads", dstName)

			if err := c.SaveFile(fh, dst); err != nil {
				log.Printf("File save error: %v", err)
				return c.Status(500).JSON(fiber.Map{"error": "failed to save file: " + err.Error()})
			}

			log.Printf("File saved: %s -> %s", fh.Filename, dst)

			attachments = append(attachments, models.Attachment{
				FileName:   fh.Filename,
				FileURL:    "/uploads/" + dstName,
				FileType:   fh.Header.Get("Content-Type"),
				UploadedAt: time.Now(),
			})
		}

		log.Printf("Processed %d files", len(attachments))

	} else {
		var req models.UpdateAchievementAttachmentsRequest
		if err := c.BodyParser(&req); err != nil {
			log.Printf("Body parse error: %v", err)
			return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
		}

		if len(req.Attachments) == 0 {
			log.Printf("JSON attachments are nil or empty")
			return c.Status(400).JSON(fiber.Map{"error": "attachments missing or null"})
		}

		attachments = req.Attachments
		log.Printf("Parsed %d attachments from JSON", len(attachments))
	}

	if len(attachments) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "no attachments to update"})
	}

	// Prevent overwriting existing attachments with nil
	res, err := s.mongoRepo.UpdateAttachments(ctx, id, attachments)
	if err != nil {
		log.Printf("UpdateAttachments repo error: %v", err)
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
