package routes

import (
	"github.com/bloomingFlower/blog-backend/controller"
	"github.com/bloomingFlower/blog-backend/middleware"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	app.Post("/api/log", controller.SaveAPILog)

	app.Post("/api/register", controller.Register)
	app.Delete("/api/user", middleware.IsAuthenticate, controller.DeleteUser)
	app.Put("/api/user", middleware.IsAuthenticate, controller.UpdateUser)
	app.Post("/api/login", controller.Login)
	app.Static("/api/uploads", "./uploads")
	app.Get("/api/posts", controller.AllPost)
	app.Get("/api/posts/search", controller.SearchPost)
	app.Put("/api/post/:id/hide", controller.HidePost)

	app.Get("/api/post/:id", controller.DetailPost)
	app.Get("/api/token", controller.GenerateToken)

	app.Use("/api", middleware.IsAuthenticate)
	app.Post("/api/posts", middleware.IsAuthenticate, controller.CreatePost)

	app.Put("/api/post/:id", middleware.IsAuthenticate, controller.UpdatePost)
	app.Get("/api/unique-post", controller.UniquePost)
	app.Delete("/api/delete-post/:id", controller.DeletePost)

	app.Post("/api/upload-file", middleware.IsAuthenticate, controller.UploadFile)
	app.Post("/api/upload-img", middleware.IsAuthenticate, controller.UploadImage)
}
