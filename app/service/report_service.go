package service

import (
	"achievement_backend/app/repository"
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
)

type ReportService struct {
	refRepo      repository.AchievementReferenceRepository
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
	mongoRepo    repository.MongoAchievementRepository
	userRepo     repository.UserRepository
}

func NewReportService(
	refRepo repository.AchievementReferenceRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	mongoRepo repository.MongoAchievementRepository,
	userRepo repository.UserRepository,
) *ReportService {
	return &ReportService{
		refRepo:      refRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
		mongoRepo:    mongoRepo,
		userRepo:     userRepo,
	}
}

type StatisticsData struct {
	Total       int64 `json:"total"`
	Verified    int64 `json:"verified"`
	Rejected    int64 `json:"rejected"`
	Pending     int64 `json:"pending"`
	Draft       int64 `json:"draft"`
	TotalPoints int64 `json:"total_points"`
}

type StudentReportDetail struct {
	StudentID      string  `json:"student_id"`
	StudentName    string  `json:"student_name"`
	Total          int64   `json:"total"`
	Verified       int64   `json:"verified"`
	Rejected       int64   `json:"rejected"`
	Pending        int64   `json:"pending"`
	Draft          int64   `json:"draft"`
	TotalPoints    int64   `json:"total_points"`
	VerifiedPoints int64   `json:"verified_points"`
	RejectionRate  float64 `json:"rejection_rate"`
	CompletionRate float64 `json:"completion_rate"`
}

// ================= GET STATISTICS =================
// GetStatistics returns overall analytics
func (s *ReportService) GetStatistics(c *fiber.Ctx) error {
	log.Printf("[REPORT] Getting overall statistics")

	// Get all achievement references
	allRefs, err := s.refRepo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to fetch statistics",
			"detail": err.Error(),
		})
	}

	stats := StatisticsData{}
	totalPoints := int64(0)

	// Count statuses
	for _, ref := range allRefs {
		stats.Total++

		switch ref.Status {
		case "verified":
			stats.Verified++
		case "rejected":
			stats.Rejected++
		case "submitted":
			stats.Pending++
		case "draft":
			stats.Draft++
		}

		// Get MongoDB detail for points
		if mongoDoc, err := s.mongoRepo.GetByID(c.Context(), ref.MongoAchievementID); err == nil {
			if mongoDoc != nil && mongoDoc.Points != nil && *mongoDoc.Points > 0 {
				totalPoints += int64(*mongoDoc.Points)
			}
		}
	}

	stats.TotalPoints = totalPoints

	log.Printf("[REPORT] Statistics: Total=%d, Verified=%d, Rejected=%d, Pending=%d, Draft=%d, TotalPoints=%d",
		stats.Total, stats.Verified, stats.Rejected, stats.Pending, stats.Draft, stats.TotalPoints)

	return c.JSON(fiber.Map{
		"data":    stats,
		"success": true,
	})
}

// ================= GET STUDENT REPORT =================
// GetStudentReport returns detailed report for a specific student
func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
	studentID := c.Params("id")

	if studentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "student id is required",
		})
	}

	log.Printf("[REPORT] Getting student report for: %s", studentID)

	// Get student details
	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "student not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to fetch student",
			"detail": err.Error(),
		})
	}

	// Get all achievement references for this student
	refs, err := s.refRepo.GetByStudentID(studentID)
	if err != nil && err != sql.ErrNoRows {
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed to fetch achievements",
			"detail": err.Error(),
		})
	}

	// Get user info for name
	user, err := s.userRepo.GetByID(student.UserID)
	if err != nil {
		log.Printf("[REPORT] Warning: could not fetch user %s: %v", student.UserID, err)
	}

	report := StudentReportDetail{
		StudentID:   student.StudentID,
		StudentName: "",
	}
	if user != nil {
		report.StudentName = user.FullName
	}

	verifiedPoints := int64(0)
	totalPoints := int64(0)

	// Count statuses and calculate points
	for _, ref := range refs {
		report.Total++

		switch ref.Status {
		case "verified":
			report.Verified++
		case "rejected":
			report.Rejected++
		case "submitted":
			report.Pending++
		case "draft":
			report.Draft++
		}

		// Get MongoDB detail for points
		if mongoDoc, err := s.mongoRepo.GetByID(c.Context(), ref.MongoAchievementID); err == nil {
			if mongoDoc != nil && mongoDoc.Points != nil && *mongoDoc.Points > 0 {
				totalPoints += int64(*mongoDoc.Points)
				if ref.Status == "verified" {
					verifiedPoints += int64(*mongoDoc.Points)
				}
			}
		}
	}

	report.TotalPoints = totalPoints
	report.VerifiedPoints = verifiedPoints

	// Calculate rates
	if report.Total > 0 {
		report.RejectionRate = float64(report.Rejected) / float64(report.Total) * 100
		report.CompletionRate = float64(report.Verified) / float64(report.Total) * 100
	}

	log.Printf("[REPORT] Student %s report: Total=%d, Verified=%d, Rejected=%d, Points=%d/%d",
		studentID, report.Total, report.Verified, report.Rejected, report.VerifiedPoints, report.TotalPoints)

	return c.JSON(fiber.Map{
		"data":    report,
		"success": true,
	})
}
