package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	database.Connect()
	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
	port := os.Getenv("PORT")
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     "http://localhost:8080, https://blog.yourrubber.duckdns.org:443",
		AllowMethods:     "POST, GET, OPTIONS, PUT, DELETE",
		AllowHeaders:     "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	}))
	baseURL := os.Getenv("BASE_URL") // Get the base URL from environment variables
	log.Printf("baseURL: %s", baseURL)
	// API 라우트 체크
	app.Get("/api/data", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Go Fiber!")
	})

	routes.Setup(app)
	// Log warning if unable to start the server
	if err := app.Listen(":" + port); err != nil {
		log.Printf("Warning: Unable to start server: %v", err)
	}
}
