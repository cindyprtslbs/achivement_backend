package service

import (
	"context"
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

func isAdmin(c *fiber.Ctx) bool {
	return c.Locals("role_name") == "Admin"
}

// =====================================================
// POINT CALCULATOR (AUTO POINTS)
// =====================================================
func CalculatePoints(req *models.CreateAchievementRequest) int {
	// Competition scoring
	if req.AchievementType == "competition" {
		level := ""
		if req.Details.CompetitionLevel != nil {
			level = *req.Details.CompetitionLevel
		}
		rank := 0
		if req.Details.Rank != nil {
			rank = *req.Details.Rank
		}

		switch level {
		case "international":
			if rank == 1 {
				return 100
			}
			if rank == 2 {
				return 80
			}
			if rank == 3 {
				return 60
			}
			return 40
		case "national":
			if rank == 1 {
				return 80
			}
			if rank == 2 {
				return 60
			}
			if rank == 3 {
				return 40
			}
			return 20
		case "regional":
			return 10
		case "local":
			return 5
		}
	}

	// Publication
	if req.AchievementType == "publication" {
		return 40
	}

	// Certification
	if req.AchievementType == "certification" {
		return 20
	}

	// Default
	return 10
}

// constructor unchanged
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

// -----------------------
// ListByRole, GetDetail, UpdateStatus, GetByStudent
// keep same as before (unchanged logic)
// -----------------------

func (s *AchievementMongoService) ListByRole(c *fiber.Ctx) error {
    userID := c.Locals("user_id")
    roleName := c.Locals("role_name")
    if userID == nil || roleName == nil {
        return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
    }

    uid := userID.(string)
    role := roleName.(string)
    ctx := c.Context()

    page := c.QueryInt("page", 1)
    limit := c.QueryInt("limit", 10)
    offset := (page - 1) * limit

    switch role {

    // ======================= ADMIN ===========================
    case "Admin":
        refs, total, err := s.refRepo.GetAllWithPagination(limit, offset)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "failed to fetch references"})
        }

        mongoIDs := []string{}
        for _, r := range refs {
            mongoIDs = append(mongoIDs, r.MongoAchievementID)
        }

        mDetails, _ := s.mongoRepo.GetManyByIDs(ctx, mongoIDs)

        out := []fiber.Map{}
        for _, r := range refs {
            out = append(out, fiber.Map{
                "reference": r,
                "detail":    mDetails[r.MongoAchievementID],
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

    // ======================= DOSEN WALI ===========================
    case "Dosen Wali":
        lecturer, err := s.lecturerRepo.GetByUserID(uid)
        if err != nil || lecturer == nil {
            return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
        }

        students, err := s.studentRepo.GetByAdvisorID(lecturer.ID)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "failed to load advisees"})
        }

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

        studentIDs := []string{}
        for _, st := range students {
            studentIDs = append(studentIDs, st.ID)
        }

        refs, total, err := s.refRepo.GetByAdviseesWithPagination(studentIDs, limit, offset)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievements"})
        }

        mongoIDs := []string{}
        for _, r := range refs {
            mongoIDs = append(mongoIDs, r.MongoAchievementID)
        }

        mDetails, _ := s.mongoRepo.GetManyByIDs(ctx, mongoIDs)

        out := []fiber.Map{}
        for _, r := range refs {
            out = append(out, fiber.Map{
                "reference": r,
                "detail":    mDetails[r.MongoAchievementID],
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

    // ======================= MAHASISWA ===========================
    case "Mahasiswa":
        student, err := s.studentRepo.GetByUserID(uid)
        if err != nil || student == nil {
            return c.Status(404).JSON(fiber.Map{"error": "student not found"})
        }

        refs, err := s.refRepo.GetByStudentID(student.ID)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievements"})
        }

        mongoIDs := []string{}
        for _, r := range refs {
            mongoIDs = append(mongoIDs, r.MongoAchievementID)
        }

        mDetails, _ := s.mongoRepo.GetManyByIDs(ctx, mongoIDs)

        out := []fiber.Map{}
        for _, r := range refs {
            out = append(out, fiber.Map{
                "reference": r,
                "detail":    mDetails[r.MongoAchievementID],
            })
        }

        return c.JSON(fiber.Map{"success": true, "data": out})
    }

    return c.Status(403).JSON(fiber.Map{"error": "invalid role"})
}


func (s *AchievementMongoService) GetDetail(c *fiber.Ctx) error {
	mongoID := c.Params("id")
	ctx := c.Context()

	userID := c.Locals("user_id")
	roleName := c.Locals("role_name")

	if userID == nil || roleName == nil {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "unauthorized"})
	}

	uid := userID.(string)
	role := roleName.(string)

	// ===== Ambil reference (WAJIB untuk RBAC) =====
	ref, err := s.refRepo.GetByMongoAchievementID(mongoID)
	if err != nil || ref == nil {
		return c.Status(404).
			JSON(fiber.Map{"error": "achievement not found"})
	}

	// ===== RBAC CHECK =====
	switch role {

	case "Admin":
		// full access

	case "Mahasiswa":
		student, err := s.studentRepo.GetByUserID(uid)
		if err != nil || student == nil {
			return c.Status(404).
				JSON(fiber.Map{"error": "student not found"})
		}

		if ref.StudentID != student.ID {
			return c.Status(fiber.StatusForbidden).
				JSON(fiber.Map{"error": "access denied"})
		}

	case "Dosen Wali":
		lecturer, err := s.lecturerRepo.GetByUserID(uid)
		if err != nil || lecturer == nil {
			return c.Status(404).
				JSON(fiber.Map{"error": "lecturer not found"})
		}

		student, err := s.studentRepo.GetByID(ref.StudentID)
		if err != nil || student == nil {
			return c.Status(404).
				JSON(fiber.Map{"error": "student not found"})
		}

		if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return c.Status(fiber.StatusForbidden).
				JSON(fiber.Map{"error": "access denied"})
		}

	default:
		return c.Status(fiber.StatusForbidden).
			JSON(fiber.Map{"error": "invalid role"})
	}

	// ===== Ambil detail dari MongoDB =====
	item, err := s.mongoRepo.GetByID(ctx, mongoID)
	if err != nil {
		log.Printf("[GetDetail] mongoRepo.GetByID error: %v", err)
		return c.Status(500).
			JSON(fiber.Map{"error": "failed to fetch achievement detail"})
	}

	if item == nil {
		return c.Status(404).
			JSON(fiber.Map{"error": "achievement not found"})
	}

	if item.IsDeleted {
		return c.Status(410).
			JSON(fiber.Map{"error": "achievement deleted"})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"reference": ref,
		"detail":    item,
	})
}


// ===========================================================
// CREATE DRAFT  (AUTO POINTS) — FIXED: don't set req.Points
// ===========================================================
func (s *AchievementMongoService) CreateDraft(c *fiber.Ctx) error {
	var req models.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	uid := c.Locals("user_id").(string)
	ctx := c.Context()

	var studentID string

	if isAdmin(c) {
		if req.StudentID == nil {
			return c.Status(400).JSON(fiber.Map{"error": "student_id required"})
		}
		studentID = *req.StudentID
	} else {
		student, _ := s.studentRepo.GetByUserID(uid)
		if student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "student not found"})
		}
		studentID = student.ID
	}

	points := CalculatePoints(&req)

	created, err := s.mongoRepo.CreateDraft(ctx, studentID, &req, points)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create"})
	}

	ref, _ := s.refRepo.Create(studentID, created.ID.Hex())

	return c.Status(201).JSON(fiber.Map{
		"success":   true,
		"detail":    created,
		"reference": ref,
	})
}

// ===========================================================
// UPDATE DRAFT (AUTO RECALCULATE POINTS) — FIXED
// ===========================================================
func (s *AchievementMongoService) UpdateDraft(c *fiber.Ctx) error {
	id := c.Params("id")
	role := c.Locals("role_name").(string)
	uid := c.Locals("user_id").(string)

	var req models.UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	item, err := s.mongoRepo.GetByID(c.Context(), id)
	if err != nil || item == nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}

	// ===== hanya draft =====
	if item.Status != models.StatusDraft {
		return c.Status(400).JSON(fiber.Map{"error": "only draft can be updated"})
	}

	// ===== RBAC =====
	switch role {

	case "Admin":
		// full akses → lanjut

	case "Mahasiswa":
		student, err := s.studentRepo.GetByUserID(uid)
		if err != nil || student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "student not found"})
		}
		if item.StudentID != student.ID {
			return c.Status(403).JSON(fiber.Map{"error": "not owner"})
		}

	default:
		return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
	}

	// ===== hitung ulang points =====
	recalc := models.CreateAchievementRequest{
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Attachments:     req.Attachments,
		Tags:            req.Tags,
	}
	points := CalculatePoints(&recalc)

	updated, err := s.mongoRepo.UpdateDraft(c.Context(), id, &req, points)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    updated,
	})
}

// ===========================================================
// SOFT DELETE
// ===========================================================
func (s *AchievementMongoService) SoftDelete(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := c.Context()

	role := c.Locals("role_name").(string)
	userID := c.Locals("user_id").(string)

	item, err := s.mongoRepo.GetByID(ctx, id)
	if err != nil || item == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement not found",
		})
	}

	// ================= ADMIN =================
	if role == "Admin" {
		goto DELETE
	}

	// ================= MAHASISWA =================
	if role == "Mahasiswa" {
		student, err := s.studentRepo.GetByUserID(userID)
		if err != nil || student == nil {
			return c.Status(403).JSON(fiber.Map{
				"error": "student not found",
			})
		}

		if item.StudentID != student.ID {
			return c.Status(403).JSON(fiber.Map{
				"error": "forbidden: not your achievement",
			})
		}

		goto DELETE
	}

	// ================= OTHERS (DOSEN WALI, DLL) =================
	return c.Status(403).JSON(fiber.Map{
		"error": "forbidden: role not allowed",
	})

DELETE:
	// Soft delete Mongo
	if err := s.mongoRepo.SoftDelete(ctx, id); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to delete achievement",
		})
	}

	ref, err := s.refRepo.GetByMongoAchievementID(id)
	if err == nil && ref != nil {
		_ = s.refRepo.SoftDelete(ref.ID)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "achievement deleted successfully",
	})
}

// ===========================================================
// UPDATE ATTACHMENTS (unchanged except no points here)
// ===========================================================
func (s *AchievementMongoService) UpdateAttachments(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := c.Context()

	role := c.Locals("role_name").(string)
	uid := c.Locals("user_id").(string)

	// ===== ambil achievement =====
	item, err := s.mongoRepo.GetByID(ctx, id)
	if err != nil || item == nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}

	// ===== RBAC =====
	switch role {

	// ================= ADMIN FULL ACCESS =================
	case "Admin":
		// langsung lanjut, tanpa cek owner

	// ================= MAHASISWA =================
	case "Mahasiswa":
		student, err := s.studentRepo.GetByUserID(uid)
		if err != nil || student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "student not found"})
		}
		if item.StudentID != student.ID {
			return c.Status(403).JSON(fiber.Map{"error": "not owner"})
		}

	// ================= ROLE LAIN =================
	default:
		return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
	}

	// ===== hanya draft yang boleh diubah =====
	if item.Status != models.StatusDraft {
		return c.Status(400).JSON(fiber.Map{"error": "only draft achievement can be updated"})
	}

	// ===== proses attachments =====
	var attachments []models.Attachment

	contentType := c.Get("Content-Type")
	isMultipart := c.Is("multipart/form-data")

	if isMultipart || (contentType != "" && len(contentType) >= 19 && contentType[:19] == "multipart/form-data") {

		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid multipart form"})
		}

		var files []*multipart.FileHeader
		for _, key := range []string{"attachments", "file", "files"} {
			if f, ok := form.File[key]; ok {
				files = f
				break
			}
		}

		for _, fh := range files {
			dstName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fh.Filename))
			dst := filepath.Join("uploads", dstName)
			_ = c.SaveFile(fh, dst)

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
			return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
		}
		attachments = req.Attachments
	}

	// ===== update mongo =====
	res, err := s.mongoRepo.UpdateAttachments(ctx, id, attachments)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

// ===========================================================
// GET BY STUDENT
// ===========================================================
func (s *AchievementMongoService) GetByStudent(c *fiber.Ctx) error {
	studentID := c.Params("id")
	ctx := c.Context()
	role := c.Locals("role_name").(string)
	userID := c.Locals("user_id").(string)

	// ================= ADMIN =================
	if role != "Admin" && role != "Dosen Wali" {
		return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
	}

	// ================= DOSEN WALI =================
	if role == "Dosen Wali" {
		lecturer, err := s.lecturerRepo.GetByUserID(userID)
		if err != nil || lecturer == nil {
			return c.Status(403).JSON(fiber.Map{"error": "lecturer not found"})
		}

		student, err := s.studentRepo.GetByID(studentID)
		if err != nil || student == nil {
			return c.Status(404).JSON(fiber.Map{"error": "student not found"})
		}

		if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
		}
	}

	// ================= FETCH DATA =================
	refs, err := s.refRepo.GetByStudentID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch references"})
	}

	mongoIDs := make([]string, 0, len(refs))
	for _, r := range refs {
		mongoIDs = append(mongoIDs, r.MongoAchievementID)
	}

	mDetails, err := s.mongoRepo.GetManyByIDs(ctx, mongoIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch achievements"})
	}

	out := []fiber.Map{}
	for _, ref := range refs {
		out = append(out, fiber.Map{
			"reference": ref,
			"detail":    mDetails[ref.MongoAchievementID],
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    out,
	})
}

// ===========================================================
// UPDATE STATUS (wrapper)
// ===========================================================
func (s *AchievementMongoService) UpdateStatus(ctx context.Context, mongoID string, status string) error {
	return s.mongoRepo.UpdateStatus(ctx, mongoID, status)
}
