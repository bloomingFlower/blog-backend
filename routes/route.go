package routes

import (
	"github.com/bloomingFlower/blog-backend/controller"
	"github.com/bloomingFlower/blog-backend/middleware"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	app.Post("/api/register", controller.Register)
	app.Post("/api/login", controller.Login)
	app.Use("/api", middleware.IsAuthenticate)
	app.Post("/create-post", controller.CreatePost)
	app.Get("/api/all-post", controller.AllPost)

}
