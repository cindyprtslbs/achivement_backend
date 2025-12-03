package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"achievement_backend/app/repository"
	"achievement_backend/app/service"
	"achievement_backend/database"
	"achievement_backend/route"
)

func main() {

	// ============================================================
	// 1. CONNECT DATABASES
	// ============================================================
	database.ConnectPostgre()
	database.ConnectMongo()

	log.Println("Database connected")

	// ============================================================
	// 2. INIT REPOSITORIES
	// ============================================================
	userRepo := repository.NewUserRepository(database.PostgreDB)
	roleRepo := repository.NewRoleRepository(database.PostgreDB)
	permissionRepo := repository.NewPermissionRepository(database.PostgreDB)
	rolePermissionRepo := repository.NewRolePermissionRepository(database.PostgreDB)

	studentRepo := repository.NewStudentRepository(database.PostgreDB)
	lecturerRepo := repository.NewLecturerRepository(database.PostgreDB)
	achievementRefRepo := repository.NewAchievementReferenceRepository(database.PostgreDB)

	achievementMongoRepo := repository.NewMongoAchievementRepository(database.MongoDB)

	// ============================================================
	// 3. INIT SERVICES
	// ============================================================
	authService := service.NewAuthService(userRepo, roleRepo, rolePermissionRepo)
	userService := service.NewUserService(userRepo, roleRepo)
	roleService := service.NewRoleService(roleRepo)
	permissionService := service.NewPermissionService(permissionRepo)
	rolePermissionService := service.NewRolePermissionService(rolePermissionRepo)

	studentService := service.NewStudentService(studentRepo, userRepo, lecturerRepo)

	// FIXED
	lecturerService := service.NewLecturerServiceWithDependencies(lecturerRepo, studentRepo, userRepo, achievementRefRepo, achievementMongoRepo)

	// FIXED
	achievementService := service.NewAchievementMongoService(
		achievementMongoRepo,
		achievementRefRepo,
		studentRepo,
		lecturerRepo,
	)

	achievementRefService := service.NewAchievementReferenceService(achievementRefRepo, achievementMongoRepo)

	achievementHistoryService := service.NewAchievementHistoryService(achievementRefRepo, achievementMongoRepo)

	reportService := service.NewReportService(achievementRefRepo, studentRepo, lecturerRepo, achievementMongoRepo, userRepo)

	// ============================================================
	// 4. INIT FIBER
	// ============================================================
	app := fiber.New()

	// Serve uploaded files statically
	app.Static("/uploads", "./uploads")

	// ============================================================
	// 5. SETUP ROUTES (TANPA PARAMETER RBAC)
	// ============================================================
	route.SetupRoutes(
		app,
		authService,
		userService,
		roleService,
		permissionService,
		rolePermissionService,
		studentService,
		lecturerService,
		achievementService,
		achievementRefService,
		achievementHistoryService,
		reportService,
	)

	// ============================================================
	// 6. START SERVER
	// ============================================================
	log.Println("Server berjalan di port 8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
