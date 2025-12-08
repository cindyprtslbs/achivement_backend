package service

import (
	models "achievement_backend/app/model"
	"achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type LecturerService struct {
	repo        repository.LecturerRepository
	studentRepo repository.StudentRepository
	userRepo    repository.UserRepository 
	refRepo     repository.AchievementReferenceRepository
	mongoRepo   repository.MongoAchievementRepository
}

func NewLecturerService(r repository.LecturerRepository, s repository.StudentRepository, u repository.UserRepository,) *LecturerService {
	return &LecturerService{
		repo:        r,
		studentRepo: s,
		userRepo:    u,
	}
}

func NewLecturerServiceWithDependencies(
    r repository.LecturerRepository,
    s repository.StudentRepository,
    u repository.UserRepository,
    ref repository.AchievementReferenceRepository,
    mongo repository.MongoAchievementRepository,
) *LecturerService {
    return &LecturerService{
        repo:        r,
        studentRepo: s,
        userRepo:    u,     
        refRepo:     ref,
        mongoRepo:   mongo,
    }
}

// GET ALL LECTURERS
func (s *LecturerService) GetAll(c *fiber.Ctx) error {
    role := c.Locals("role_name")
    userID := c.Locals("user_id")

    if role == nil || userID == nil {
        return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
    }

    r := role.(string)
    uid := userID.(string)

    switch r {

    case "Admin":
        // Admin melihat semua dosen
        data, err := s.repo.GetAll()
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "failed to get lecturers"})
        }
        return c.JSON(fiber.Map{
            "success": true,
            "data":    data,
        })

    case "Dosen Wali":
        // Dosen Wali hanya melihat dirinya sendiri
        lecturer, err := s.repo.GetByUserID(uid)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{
                "error": "failed to fetch lecturer profile",
            })
        }
        if lecturer == nil {
            return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
        }

        return c.JSON(fiber.Map{
            "success": true,
            "data":    []models.Lecturer{*lecturer}, 
        })

    default:
        return c.Status(403).JSON(fiber.Map{
            "error": "forbidden: role cannot access lecturer list",
        })
    }
}


// GET LECTURER BY USER ID
func (s *LecturerService) GetByUserID(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	lecturer, err := s.repo.GetByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    lecturer,
	})
}

// CREATE LECTURER (ADMIN)
func (s *LecturerService) SetLecturerProfile(c *fiber.Ctx) error {
	userId := c.Params("id")

	// Parse request
	var req models.SetLecturerProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// 1. Check user exists
	user, err := s.userRepo.GetByID(userId)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	// 2. Check role must be Dosen Wali
	if user.RoleName != "Dosen Wali" {
		return fiber.NewError(fiber.StatusBadRequest, "User is not assigned as Dosen Wali")
	}

	// 3. Check if lecturer profile exists
	existing, err := s.repo.GetByUserID(userId)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	// 4A. Update existing profile
	if existing != nil {
		updated, err := s.repo.Update(existing.ID, models.UpdateLecturerRequest{
			LecturerID: req.LecturerID,
			Department: req.Department,
		})
		if err != nil {
			return fiber.ErrInternalServerError
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Lecturer profile updated successfully",
			"data":    updated,
		})
	}

	// 4B. Create new profile
	created, err := s.repo.Create(models.CreateLecturerRequest{
		UserID:     userId,
		LecturerID: req.LecturerID,
		Department: req.Department,
	})
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Lecturer profile created successfully",
		"data":    created,
	})
}

// UPDATE LECTURER (ADMIN)
func (s *LecturerService) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateLecturerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	lecturer, err := s.repo.Update(id, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update lecturer"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "lecturer updated",
		"data":    lecturer,
	})
}

// ============================================
// GET ADVISEES (Mahasiswa Bimbingan Dosen)
// GET /api/v1/lecturers/:id/advisees
// ============================================
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	lecturerID := c.Params("id")

	students, err := s.studentRepo.GetByAdvisorID(lecturerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get advisees"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    students,
	})
}
