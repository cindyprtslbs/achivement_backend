package route

// import (
// 	"crud-app/app/repository"
// 	"crud-app/app/service"
// 	"crud-app/middleware"

// 	"github.com/gofiber/fiber/v2"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// func SetupRoutes(app *fiber.App, db *mongo.Database) {
// 	// -------------------------
// 	// Base groups
// 	// -------------------------
// 	api := app.Group("/api")
// 	unair := app.Group("/unair")

// 	// =========================
// 	// AUTH ROUTES
// 	// =========================
// 	authRepo := repository.NewAuthRepository(db)
// 	authService := service.NewAuthService(authRepo)

// 	api.Post("/login", authService.Login)

// 	// Protected route (harus login)
// 	protected := api.Group("", middleware.AuthRequired())
// 	protected.Get("/profile", authService.GetProfile)

// 	// =========================
// 	// ALUMNI ROUTES
// 	// =========================
// 	alumniRepo := repository.NewAlumniRepository(db)
// 	alumniService := service.NewAlumniService(alumniRepo)

// 	alumni := unair.Group("/alumni")
// 	alumni.Get("/", alumniService.GetAlumniService)
// 	alumni.Get("/without-pekerjaan", middleware.AuthRequired(), alumniService.GetWithoutPekerjaan)
// 	alumni.Get("/:id", middleware.AuthRequired(), alumniService.GetByID)

// 	alumni.Post("/", middleware.AuthRequired(), middleware.AdminOnly(), alumniService.Create)
// 	alumni.Put("/:id", middleware.AuthRequired(), middleware.AdminOnly(), alumniService.Update)
// 	alumni.Delete("/:id", middleware.AuthRequired(), alumniService.SoftDelete)
// 	alumni.Patch("/:id", middleware.AuthRequired(), alumniService.Restore)

// 	// =========================
// 	// PEKERJAAN ALUMNI ROUTES
// 	// =========================
// 	pekerjaanRepo := repository.NewPekerjaanRepository(db)
// 	pekerjaanService := service.NewPekerjaanService(pekerjaanRepo)

// 	pekerjaan := unair.Group("/pekerjaan-alumni")
// 	pekerjaan.Get("/", pekerjaanService.GetPekerjaanService)
// 	pekerjaan.Get("/trash", middleware.AuthRequired(), pekerjaanService.GetTrash)
// 	pekerjaan.Get("/:id", middleware.AuthRequired(), pekerjaanService.GetByID)
// 	pekerjaan.Get("/alumni/:alumni_id", middleware.AuthRequired(), middleware.AdminOnly(), pekerjaanService.GetByAlumniID)

// 	pekerjaan.Post("/", middleware.AuthRequired(), middleware.AdminOnly(), pekerjaanService.Create)
// 	pekerjaan.Put("/:id", middleware.AuthRequired(), middleware.AdminOnly(), pekerjaanService.Update)
// 	pekerjaan.Delete("/:id", middleware.AuthRequired(), pekerjaanService.SoftDelete)
// 	pekerjaan.Patch("/:id", middleware.AuthRequired(), pekerjaanService.Restore)

// 	// Opsional tambahan untuk restore dan hard delete
// 	pekerjaan.Put("/restore/:id", middleware.AuthRequired(), pekerjaanService.Restore)
// 	pekerjaan.Delete("/trash/delete/:id", middleware.AuthRequired(), pekerjaanService.Delete)

// 	// =========================
// 	// UPLOAD FILES ROUTES
// 	// =========================
// 	files := app.Group("/files")
	
// 	fileRepo := repository.NewFileRepository(db)
// 	uploadPath := "./uploads"
// 	fileService := service.NewFileService(fileRepo, uploadPath)

// 	files.Post("/upload/foto", middleware.AuthRequired(), fileService.UploadFoto)
// 	files.Post("/upload/sertifikat", middleware.AuthRequired(), fileService.UploadSertifikat)
// 	files.Get("/", middleware.AuthRequired(), fileService.GetAllFiles)
// 	files.Get("/:id", middleware.AuthRequired(), fileService.GetFileByID)
// 	files.Delete("/:id", middleware.AuthRequired(), fileService.DeleteFile)
// }
