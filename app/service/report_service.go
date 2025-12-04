package service

import (
    "sort"

    model "achievement_backend/app/model"
    "achievement_backend/app/repository"

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

type StatisticsOutput struct {
	Total             int64            `json:"total"`
	PerType           map[string]int64 `json:"per_type"`
	PerMonth          map[string]int64 `json:"per_month"`
	CompetitionLevels map[string]int64 `json:"competition_levels"`
	TopStudents       []TopStudent     `json:"top_students"`
}

type TopStudent struct {
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	Points    int64  `json:"points"`
}

func (s *ReportService) GetStatistics(c *fiber.Ctx) error {
	role := c.Locals("role_name").(string)
	userID := c.Locals("user_id").(string)

	var refs []model.AchievementReference
	var total int64
	var err error

	limit := 1_000_000
	offset := 0

	// ADMIN — full access
	if role == "Admin" {
		refs, total, err = s.refRepo.GetAllWithPagination(limit, offset)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch data"})
		}
	}

	// DOSEN WALI — only advisees
	if role == "Dosen Wali" {
		lect, _ := s.lecturerRepo.GetByUserID(userID)
		if lect == nil {
			return c.Status(403).JSON(fiber.Map{"error": "lecturer not found"})
		}

		advisees, _ := s.studentRepo.GetByAdvisorID(lect.ID)
		ids := []string{}
		for _, s := range advisees {
			ids = append(ids, s.ID)
		}

		refs, total, err = s.refRepo.GetByAdviseesWithPagination(ids, limit, offset)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch data"})
		}
	}

	// MAHASISWA — only own achievement
	if role == "Mahasiswa" {
		stu, _ := s.studentRepo.GetByUserID(userID)
		if stu == nil {
			return c.Status(403).JSON(fiber.Map{"error": "student not found"})
		}

		refs, err = s.refRepo.GetByStudentID(stu.ID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch data"})
		}
		total = int64(len(refs))
	}

	// ——————————————————
	// BUILD STATISTICS
	// ——————————————————
	output := StatisticsOutput{
		Total:             total,
		PerType:           map[string]int64{},
		PerMonth:          map[string]int64{},
		CompetitionLevels: map[string]int64{},
		TopStudents:       []TopStudent{},
	}

	pointsMap := map[string]int64{}

	for _, ref := range refs {
		mg, err := s.mongoRepo.GetByID(c.Context(), ref.MongoAchievementID)
		if err != nil || mg == nil {
			continue
		}

		output.PerType[mg.AchievementType]++

		month := ref.CreatedAt.Format("2006-01")
		output.PerMonth[month]++

		level := "unknown"
		if mg.Details.CompetitionLevel != nil {
			level = *mg.Details.CompetitionLevel
		}
		output.CompetitionLevels[level]++

		if mg.Points != nil {
			pointsMap[ref.StudentID] += int64(*mg.Points)
		}
	}

	// ranking
	type kv struct {
		ID     string
		Points int64
	}

	var sorted []kv
	for id, pts := range pointsMap {
		sorted = append(sorted, kv{id, pts})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Points > sorted[j].Points
	})

	for i, st := range sorted {
		if i >= 10 {
			break
		}

		stuObj, _ := s.studentRepo.GetByID(st.ID)
		user, _ := s.userRepo.GetByID(stuObj.UserID)

		name := "Unknown"
		if user != nil {
			name = user.FullName
		}

		output.TopStudents = append(output.TopStudents, TopStudent{
			StudentID: st.ID,
			Name:      name,
			Points:    st.Points,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    output,
	})
}

func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
	role := c.Locals("role_name").(string)
	userID := c.Locals("user_id").(string)
	targetStudentID := c.Params("id")

	// Mahasiswa → hanya miliknya sendiri
	if role == "Mahasiswa" {
		stu, _ := s.studentRepo.GetByUserID(userID)
		if stu == nil || stu.ID != targetStudentID {
			return c.Status(403).JSON(fiber.Map{"error": "access denied"})
		}
	}

	// Dosen Wali → hanya advisee
	if role == "Dosen Wali" {
		lect, _ := s.lecturerRepo.GetByUserID(userID)
		if lect == nil {
			return c.Status(403).JSON(fiber.Map{"error": "lecturer not found"})
		}

		advisees, _ := s.studentRepo.GetByAdvisorID(lect.ID)
		allowed := false
		for _, a := range advisees {
			if a.ID == targetStudentID {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.Status(403).JSON(fiber.Map{"error": "access denied"})
		}
	}

	// Admin → skip restriction

	data, err := s.refRepo.GetByStudentID(targetStudentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch data"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}


