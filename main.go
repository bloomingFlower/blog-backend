package main

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
	"log"
	"os"

	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	database.Connect()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     "http://localhost:8080",
		AllowMethods:     "POST, GET, OPTIONS, PUT, DELETE",
		AllowHeaders:     "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	}))

	// 예시 API 라우트
	app.Get("/api/data", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Go Fiber!")
	})

	routes.Setup(app)
	err = app.Listen(":" + port)
	if err != nil {
		return
	}
}
