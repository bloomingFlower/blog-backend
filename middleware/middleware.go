package middleware

import (
	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/bloomingFlower/blog-backend/util"
	"github.com/gofiber/fiber/v2"
)

func IsAuthenticate(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")
	id, err := util.ParseJwt(cookie)
	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}

	var user models.User
	database.DB.Where("id = ?", id).First(&user)
	if user.ID == 0 {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "User not found",
		})
	}
	c.Locals("userID", id) // Set the user ID in locals

	return c.Next()
}
