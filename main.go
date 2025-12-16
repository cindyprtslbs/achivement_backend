package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	// swagger docs
	_ "achievement_backend/docs"

	"achievement_backend/app/repository"
	"achievement_backend/app/service"
	"achievement_backend/database"
	"achievement_backend/route"
)

// ============================================================
// SWAGGER METADATA
// ============================================================

// @title Sistem Pelaporan Prestasi Mahasiswa
// @version 1.0
// @description API Backend untuk Sistem Pelaporan dan Verifikasi Prestasi Mahasiswa
// @termsOfService http://swagger.io/terms/

// @contact.name Cindy Permatasari Lubis
// @contact.email cindy@example.com

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description JWT Token dengan format: Bearer <token>

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
	// permissionRepo := repository.NewPermissionRepository(database.PostgreDB)
	rolePermissionRepo := repository.NewRolePermissionRepository(database.PostgreDB)

	studentRepo := repository.NewStudentRepository(database.PostgreDB)
	lecturerRepo := repository.NewLecturerRepository(database.PostgreDB)
	achievementRefRepo := repository.NewAchievementReferenceRepository(database.PostgreDB)

	achievementMongoRepo := repository.NewMongoAchievementRepository(database.MongoDB)

	// ============================================================
	// 3. INIT SERVICES
	// ============================================================
	authService := service.NewAuthService(
		userRepo,
		roleRepo,
		rolePermissionRepo,
		studentRepo,
		lecturerRepo,
	)

	userService := service.NewUserService(
		userRepo,
		roleRepo,
		studentRepo,
		lecturerRepo,
	)

	// permissionService := service.NewPermissionService(permissionRepo)

	studentService := service.NewStudentService(
		studentRepo,
		userRepo,
		lecturerRepo,
	)

	lecturerService := service.NewLecturerServiceWithDependencies(
		lecturerRepo,
		studentRepo,
		userRepo,
		achievementRefRepo,
		achievementMongoRepo,
	)

	achievementService := service.NewAchievementMongoService(
		achievementMongoRepo,
		achievementRefRepo,
		studentRepo,
		lecturerRepo,
	)

	achievementRefService := service.NewAchievementReferenceService(
		achievementRefRepo,
		achievementMongoRepo,
		studentRepo,
		lecturerRepo,
	)

	achievementHistoryService := service.NewAchievementHistoryService(
		achievementRefRepo,
		achievementMongoRepo,
		studentRepo,
		lecturerRepo,
	)

	reportService := service.NewReportService(
		achievementRefRepo,
		studentRepo,
		lecturerRepo,
		achievementMongoRepo,
		userRepo,
	)

	// ============================================================
	// 4. INIT FIBER
	// ============================================================
	app := fiber.New()

	// Serve uploaded files
	app.Static("/uploads", "./uploads")

	// ============================================================
	// 5. SETUP ROUTES
	// ============================================================
	route.SetupRoutes(
		app,
		authService,
		userService,
		// permissionService,
		studentService,
		lecturerService,
		achievementService,
		achievementRefService,
		achievementHistoryService,
		reportService,
	)

	// ============================================================
	// 6. SWAGGER ROUTE
	// ============================================================
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Optional: redirect root ke Swagger
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html", fiber.StatusMovedPermanently)
	})

	// ============================================================
	// 7. START SERVER
	// ============================================================
	log.Println("Server berjalan di port 8080")
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
