package routes

import (
	"github.com/bloomingFlower/blog-backend/controller"
	"github.com/bloomingFlower/blog-backend/middleware"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	app.Post("/api/v1/log", controller.SaveAPILog)

	app.Post("/api/v1/register", controller.Register)
	app.Delete("/api/v1/user", middleware.IsAuthenticate, controller.DeleteUser)
	app.Put("/api/v1/user", middleware.IsAuthenticate, controller.UpdateUser)
	app.Post("/api/v1/login", controller.Login)

	app.Static("/api/v1/uploads", "./uploads")
	app.Get("/api/v1/posts", controller.AllPost)
	app.Get("/api/v1/posts/search", controller.SearchPost)
	app.Put("/api/v1/post/:id/hide", controller.HidePost)

	app.Get("/api/v1/post/:id", controller.DetailPost)

	app.Use("/api/v1", middleware.IsAuthenticate)
	app.Post("/api/v1/posts", middleware.IsAuthenticate, controller.CreatePost)

	app.Put("/api/v1/post/:id", middleware.IsAuthenticate, controller.UpdatePost)
	app.Get("/api/v1/unique-post", controller.UniquePost)
	app.Delete("/api/v1/delete-post/:id", controller.DeletePost)

	app.Post("/api/v1/upload-file", middleware.IsAuthenticate, controller.UploadFile)
	app.Post("/api/v1/upload-img", middleware.IsAuthenticate, controller.UploadImage)

	//social login
	app.Get("/api/v1/auth/google/login", controller.GoogleLogin)
	app.Get("/api/v1/auth/google/callback", controller.GoogleCallback)
	app.Get("/api/v1/auth/github/login", controller.GithubLogin)
	app.Get("/api/v1/auth/github/callback", controller.GithubCallback)
	app.Get("/api/v1/auth/metamask/login", controller.MetamaskLogin)
	//app.Get("/api/v1/auth/metamask/callback", controller.MetamaskCallback)
}
