package service

import (
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

// GET ALL LECTURERS (ADMIN ONLY)
func (s *LecturerService) GetAll(c *fiber.Ctx) error {
	role := c.Locals("role_name")

	if role == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	if role.(string) != "Admin" {
		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden: admin only",
		})
	}

	data, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get lecturers"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}


// GET LECTURER BY USER ID
func (s *LecturerService) GetByUserID(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	role := c.Locals("role_name").(string)
	uid := c.Locals("user_id").(string)

	if role == "Dosen Wali" && uid != userID {
		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden",
		})
	}

	lecturer, err := s.repo.GetByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    lecturer,
	})
}

// ============================================
// GET ADVISEES (ADMIN & DOSEN WALI)
// ============================================
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	lecturerID := c.Params("id")

	role := c.Locals("role_name")
	userID := c.Locals("user_id")

	if role == nil || userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	r := role.(string)
	uid := userID.(string)

	// ================= ADMIN =================
	if r == "Admin" {
		students, err := s.studentRepo.GetByAdvisorID(lecturerID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "failed to get advisees",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data":    students,
		})
	}

	// ================= DOSEN WALI =================
	if r == "Dosen Wali" {
		lecturer, err := s.repo.GetByUserID(uid)
		if err != nil || lecturer == nil {
			return c.Status(403).JSON(fiber.Map{
				"error": "lecturer not found",
			})
		}

		// dosen hanya boleh lihat bimbingannya sendiri
		if lecturer.ID != lecturerID {
			return c.Status(403).JSON(fiber.Map{
				"error": "forbidden: cannot access other lecturer advisees",
			})
		}

		students, err := s.studentRepo.GetByAdvisorID(lecturer.ID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "failed to get advisees",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data":    students,
		})
	}

	// ================= OTHERS =================
	return c.Status(403).JSON(fiber.Map{
		"error": "forbidden",
	})
}

