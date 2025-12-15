package route

import (
	"github.com/gofiber/fiber/v2"

	"achievement_backend/app/service"
	"achievement_backend/middleware"
)

func SetupRoutes(
	app *fiber.App,

	authService *service.AuthService,
	userService *service.UserService,
	permissionService *service.PermissionService,
	studentService *service.StudentService,
	lecturerService *service.LecturerService,
	achievementService *service.AchievementMongoService,
	achievementRefService *service.AchievementReferenceService,
	achievementHistoryService *service.AchievementHistoryService,
	reportService *service.ReportService,
) {

	api := app.Group("/api/v1")

	// AUTH
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login) // all roles
	auth.Post("/refresh", authService.RefreshToken) // all roles
	auth.Post("/logout", middleware.AuthRequired(), authService.Logout) // all roles
	auth.Get("/profile", middleware.AuthRequired(), authService.GetProfile) // all roles

	v1 := api.Use(middleware.AuthRequired())

	// USERS
	users := v1.Group("/users")
	users.Get("/", middleware.PermissionRequired("user:manage"), userService.GetAll) // only admin
	users.Get("/:id", middleware.PermissionRequired("user:manage"), userService.GetByID) // only admin
	users.Post("/", middleware.PermissionRequired("user:manage"), userService.Create) // only admin
	users.Put("/:id", middleware.PermissionRequired("user:manage"), userService.Update) // only admin
	users.Delete("/:id", middleware.PermissionRequired("user:manage"), userService.Delete) // only admin
	users.Put("/:id/password", middleware.PermissionRequired("user:manage"), userService.UpdatePassword) // only admin

	// ============= ACHIEVEMENTS (Mongo) =============
	ach := v1.Group("/achievements")

	// READ ACHIEVEMENTS
	ach.Get("/", middleware.PermissionRequired("achievement:read"), achievementService.ListByRole) // all roles
	ach.Get("/:id", middleware.PermissionRequired("achievement:read"), achievementService.GetDetail) // all roles
	ach.Get("/:id/history", middleware.PermissionRequired("achievement:read"), achievementHistoryService.GetHistory) // all roles

	// CRUD ACHIEVEEMNTS (MAHASISWA)
	ach.Post("/", middleware.PermissionRequired("achievement:create"), achievementService.CreateDraft) // only admin and student
	ach.Put("/:id", middleware.PermissionRequired("achievement:update"), achievementService.UpdateDraft) // only admin and student
	ach.Delete("/:id", middleware.PermissionRequired("achievement:delete"), achievementService.SoftDelete) // only admin and student

	// attachments
	ach.Post("/:id/attachments", middleware.PermissionRequired("achievement:update"), achievementService.UpdateAttachments) // only admin and student
	
	// Workflow
	ach.Post("/:id/submit", middleware.PermissionRequired("achievement:update"), achievementRefService.Submit) // only admin and student
	ach.Post("/:id/verify", middleware.PermissionRequired("achievement:verify"), achievementRefService.Verify) // only admin and lecturer
	ach.Post("/:id/reject", middleware.PermissionRequired("achievement:verify"), achievementRefService.Reject) // only admin and lecturer

	// REPORTS
	reports := v1.Group("/reports")
	reports.Get("/statistics", middleware.PermissionRequired("achievement:read"), reportService.GetStatistics) // all roles
	reports.Get("/student/:id", middleware.PermissionRequired("achievement:read"), reportService.GetStudentReport) // all roles

	// STUDENTS
	students := v1.Group("/students")
	students.Get("/", middleware.PermissionRequired("user:manage"), studentService.GetAll) // only admin
	students.Get("/:id", middleware.PermissionRequired("user:manage"), studentService.GetByID) // only admin
	students.Get("/:id/achievements", middleware.PermissionRequired("achievement:read"), achievementService.GetByStudent) // only admin and lecturer
	students.Put("/:id/advisor", middleware.PermissionRequired("user:manage"), studentService.UpdateAdvisor) // only admin

	// LECTURERS
	lecturers := v1.Group("/lecturers")
	lecturers.Get("/", middleware.PermissionRequired("user:manage"), lecturerService.GetAll) // only admin
	lecturers.Get("/:id/advisees", middleware.PermissionRequired("user:manage"), lecturerService.GetAdvisees) // only admin and lecturer
}