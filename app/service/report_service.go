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

// GetStatistics godoc
// @Summary Mendapatkan statistik prestasi
// @Description Mendapatkan statistik prestasi berdasarkan peran pengguna yang mengakses
// @Description Akses:
// @Description - Admin: semua data
// @Description - Mahasiswa: hanya achievement miliknya
// @Description - Dosen Wali: hanya achievement mahasiswa bimbingan
// @Tags Report
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Statistik prestasi"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil data"
// @Security Bearer
// @Router /api/v1/reports/statistics [get]
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

// GetStudentReport godoc
// @Summary Mendapatkan laporan prestasi mahasiswa
// @Description Mendapatkan laporan prestasi mahasiswa berdasarkan ID mahasiswa dengan filter role pengguna yang mengakses
// @Description Akses:
// @Description - Admin: semua data
// @Description - Mahasiswa: hanya own report
// @Description - Dosen Wali: hanya laporan mahasiswa bimbingan
// @Tags Report
// @Accept json
// @Produce json
// @Param id path string true "ID Mahasiswa"
// @Success 200 {object} map[string]interface{} "Laporan prestasi mahasiswa"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Mahasiswa tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil data"
// @Security Bearer
// @Router /api/v1/reports/student/{id} [get]
func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
	role := c.Locals("role_name").(string)
	loggedUID := c.Locals("user_id").(string)
	studentID := c.Params("id")

	// ========================
	// 1. STUDENT
	// ========================
	student, err := s.studentRepo.GetByID(studentID)
	if err != nil || student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	// ========================
	// 2. RBAC
	// ========================
	switch role {

	case "Mahasiswa":
		if student.UserID != loggedUID {
			return c.Status(403).JSON(fiber.Map{
				"error": "forbidden: access own report only",
			})
		}

	case "Dosen Wali":
		lect, err := s.lecturerRepo.GetByUserID(loggedUID)
		if err != nil || lect == nil || student.AdvisorID == nil || *student.AdvisorID != lect.ID {
			return c.Status(403).JSON(fiber.Map{
				"error": "forbidden: not your advisee",
			})
		}

	case "Admin":
		// full access

	default:
		return c.Status(403).JSON(fiber.Map{"error": "invalid role"})
	}

	// ========================
	// 3. ALL ACHIEVEMENTS (NO FILTER)
	// ========================
	refs, err := s.refRepo.GetByStudentID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch achievements",
		})
	}

	// ========================
	// 4. BUILD DATA
	// ========================
	var (
		totalPoints int64 = 0
		levels             = map[string]int{}
		detailed           = []fiber.Map{}
	)

	for _, ref := range refs {
		mg, err := s.mongoRepo.GetByID(c.Context(), ref.MongoAchievementID)
		if err != nil || mg == nil {
			continue // skip broken data
		}

		points := int64(0)
		if mg.Points != nil {
			points = int64(*mg.Points)
			totalPoints += points
		}

		level := "unknown"
		if mg.Details.CompetitionLevel != nil {
			level = *mg.Details.CompetitionLevel
		}
		levels[level]++

		detailed = append(detailed, fiber.Map{
			"id":     ref.ID,
			"title":  mg.Title,
			"type":   mg.AchievementType,
			"level":  level,
			"status": ref.Status,
			"points": points,
		})
	}

	// ========================
	// 5. RESPONSE
	// ========================
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"student": fiber.Map{
				"id":            student.ID,
				"name":          student.FullName,
				"student_id":    student.StudentID,
				"program_study": student.ProgramStudy,
				"academic_year": student.AcademicYear,
			},
			"summary": fiber.Map{
				"total_achievements": len(detailed),
				"total_points":       totalPoints,
				"competition_levels": levels,
			},
			"achievements": detailed,
		},
	})
}

