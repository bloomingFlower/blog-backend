package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"

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

	// CORS 미들웨어 설정
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     "http://localhost:8080, https://blog.yourrubber.duckdns.org:443",
		AllowMethods:     "POST, GET, OPTIONS, PUT, DELETE",
		AllowHeaders:     "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	}))

	// 전역 속도 제한 미들웨어 설정
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"message": "Rate limit exceeded",
			})
		},
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
