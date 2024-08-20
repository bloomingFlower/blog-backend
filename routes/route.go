package routes

import (
	"time"

	"github.com/bloomingFlower/blog-backend/controller"
	"github.com/bloomingFlower/blog-backend/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func Setup(app *fiber.App) {
	// API v1 그룹 생성 및 속도 제한 설정
	v1 := app.Group("/api/v1", limiter.New(limiter.Config{
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

	// 로그 및 인증 관련 라우트
	v1.Post("/log", controller.SaveAPILog)
	v1.Post("/register", controller.Register)
	v1.Post("/login", controller.Login)

	// 인증이 필요한 사용자 관련 라우트
	user := v1.Group("/user")
	user.Use(middleware.IsAuthenticate)
	user.Delete("", controller.DeleteUser)
	user.Put("", controller.UpdateUser)

	// 정적 파일 서빙
	v1.Static("/download", "./uploads")

	// 포스트 관련 라우트
	posts := v1.Group("/posts")
	posts.Get("", controller.AllPost)
	posts.Get("/search", controller.SearchPost)
	posts.Post("", middleware.IsAuthenticate, controller.CreatePost)

	post := v1.Group("/post")
	post.Get("/:id", controller.DetailPost)
	post.Put("/:id", middleware.IsAuthenticate, controller.UpdatePost)
	post.Put("/:id/hide", middleware.IsAuthenticate, controller.HidePost)
	post.Delete("/:id", middleware.IsAuthenticate, controller.DeletePost)

	v1.Get("/unique-post", controller.UniquePost)
	v1.Get("/rss", controller.RSSFeed)

	// 나는
	v1.Get("/about-me/:id", controller.GetAboutInfo)
	v1.Post("/about-me", controller.CreateAboutInfo)
	v1.Put("/about-me/:id", controller.UpdateAboutInfo)
	v1.Delete("/about-me/:id", controller.DeleteAboutInfo)

	// Section routes
	v1.Post("/about-me/:aboutInfoID/sections", controller.CreateSection)
	v1.Put("/about-me/:aboutInfoID/sections/:sectionID", controller.UpdateSection)
	v1.Delete("/about-me/:aboutInfoID/sections/:sectionID", controller.DeleteSection)

	// SectionItem routes
	v1.Post("/about-me/:aboutInfoID/sections/:sectionID/items", controller.CreateSectionItem)
	v1.Put("/about-me/:aboutInfoID/sections/:sectionID/items/:itemID", controller.UpdateSectionItem)
	v1.Delete("/about-me/:aboutInfoID/sections/:sectionID/items/:itemID", controller.DeleteSectionItem)

	// 파일 업로드 관련 라우트
	v1.Post("/upload-file", middleware.IsAuthenticate, controller.UploadFile)
	v1.Post("/upload-img", middleware.IsAuthenticate, controller.UploadImage)
	v1.Get("/files/:id/:filename", middleware.IsAuthenticate, controller.ServeFile)

	// 소셜 로그인 관련 라우트
	auth := v1.Group("/auth")
	auth.Get("/google/login", controller.GoogleLogin)
	auth.Get("/google/callback", controller.GoogleCallback)
	auth.Get("/github/login", controller.GithubLogin)
	auth.Get("/github/callback", controller.GithubCallback)
	auth.Get("/metamask/login", controller.MetamaskLogin)

	// 댓글 관련 라우트
	comments := v1.Group("/comments")
	comments.Get("/:postId", controller.GetComments)
	comments.Post("/:postId", controller.CreateComment)
	comments.Post("/:postId/:commentId/replies", controller.CreateReply)
	comments.Post("/:postId/:commentId/vote", controller.VoteComment)
}
