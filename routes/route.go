package routes

import (
	"github.com/bloomingFlower/blog-backend/controller"
	"github.com/bloomingFlower/blog-backend/middleware"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	app.Post("/api/register", controller.Register)
	app.Delete("/api/user", middleware.IsAuthenticate, controller.DeleteUser)
	app.Put("/api/user", middleware.IsAuthenticate, controller.UpdateUser)
	app.Post("/api/login", controller.Login)

	app.Use("/api", middleware.IsAuthenticate)
	app.Post("/api/posts", middleware.IsAuthenticate, controller.CreatePost)
	app.Get("/api/posts", controller.AllPost)
	app.Get("/api/post/:id", controller.DetailPost)
	app.Put("/api/post/:id", middleware.IsAuthenticate, controller.UpdatePost)
	app.Get("/api/unique-post", controller.UniquePost)
	app.Delete("/api/delete-post/:id", controller.DeletePost)

	app.Post("/api/upload-file", controller.UploadFile)
	app.Static("/api/uploads", "./uploads")
}
