package main

import (
	"log"

	"crud-app/config"
	// "crud-app/database"
	// "crud-app/route"

	_ "crud-app/docs"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title		CRUD APPLICATION
// @version		1.0
// @description API sederhana untuk operasi CRUD menggunakan Fiber dan MongoDB
// @host 		localhost:3000
// @BasePath 	/
// @schemes 	http
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description "JWT Token dengan prefix 'Bearer '"
func main() {
	config.LoadEnv()
	config.InitLogger()

	// db := database.ConnectMongo()

	app := config.NewApp()

	// route.SetupRoutes(app, db)

	// Setup Swagger documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Redirect root to Swagger UI
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html", 301)
	})

	// Jalankan server
	port := config.GetEnv("APP_PORT", "3000")
	config.Logger.Println("Server running at http://localhost:" + port)
	config.Logger.Println("Swagger UI available at http://localhost:" + port + "/swagger/index.html")
	log.Fatal(app.Listen(":" + port))
}
