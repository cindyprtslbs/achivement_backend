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
	roleService *service.RoleService,
	permissionService *service.PermissionService,
	rolePermissionService *service.RolePermissionService,
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
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.RefreshToken)
	auth.Post("/logout", middleware.AuthRequired(), authService.Logout)
	auth.Get("/profile", middleware.AuthRequired(), authService.GetProfile)

	v1 := api.Use(middleware.AuthRequired())

	// USERS
	users := v1.Group("/users")
	users.Get("/", middleware.PermissionRequired("user:manage"), userService.GetAll)
	users.Get("/:id", middleware.PermissionRequired("user:manage"), userService.GetByID)
	users.Post("/", middleware.PermissionRequired("user:manage"), userService.Create)
	users.Put("/:id", middleware.PermissionRequired("user:manage"), userService.Update)
	users.Delete("/:id", middleware.PermissionRequired("user:manage"), userService.Delete)
	users.Put("/:id/role", middleware.PermissionRequired("user:manage"), userService.UpdatePassword)

	users.Put("/:id/lecturer-profile", middleware.PermissionRequired("user:manage"), lecturerService.SetLecturerProfile)
	users.Put("/:id/student-profile", middleware.PermissionRequired("user:manage"), studentService.SetStudentProfile)

	// ============= ACHIEVEMENTS (Mongo) =============
	ach := v1.Group("/achievements")

	ach.Get("/", middleware.PermissionRequired("achievement:read"), achievementService.ListByRole)
	ach.Get("/:id", middleware.PermissionRequired("achievement:read"), achievementService.GetDetail)
	ach.Post("/", middleware.PermissionRequired("achievement:create"), achievementService.CreateDraft)
	ach.Put("/:id", middleware.PermissionRequired("achievement:update"), achievementService.UpdateDraft)
	ach.Delete("/:id", middleware.PermissionRequired("achievement:delete"), achievementService.SoftDelete)
	ach.Post("/:id/attachments", middleware.PermissionRequired("achievement:update"), achievementService.UpdateAttachments)
	ach.Get("/:id/history", middleware.PermissionRequired("achievement:read"), achievementHistoryService.GetHistory)

	// status actions
	ach.Post("/:id/submit", middleware.PermissionRequired("achievement:submit"), achievementRefService.Submit)
	ach.Post("/:id/verify", middleware.PermissionRequired("achievement:verify"), achievementRefService.Verify)
	ach.Post("/:id/reject", middleware.PermissionRequired("achievement:reject"), achievementRefService.Reject)

	// REPORTS
	reports := v1.Group("/reports")
	reports.Get("/statistics", middleware.PermissionRequired("achievement:read"), reportService.GetStatistics)
	reports.Get("/student/:id", middleware.PermissionRequired("student:read"), reportService.GetStudentReport)

	// STUDENTS
	students := v1.Group("/students")
	students.Get("/", middleware.PermissionRequired("student:read"), studentService.GetAll)
	students.Get("/:id", middleware.PermissionRequired("student:read"), studentService.GetByID)
	students.Get("/:id/achievements", middleware.PermissionRequired("student:read"), achievementService.GetByStudent)
	students.Put("/:id/advisor", middleware.PermissionRequired("student:assign_advisor"), studentService.UpdateAdvisor)

	// LECTURERS
	lecturers := v1.Group("/lecturers")
	lecturers.Get("/", middleware.PermissionRequired("lecturer:read"), lecturerService.GetAll)
	lecturers.Get("/:id/advisees", middleware.PermissionRequired("lecturer:advisees"), lecturerService.GetAdvisees)
}
