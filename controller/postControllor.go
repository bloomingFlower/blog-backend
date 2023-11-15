package controller

import (
	"fmt"
	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/gofiber/fiber/v2"
	"math"
	"strconv"
)

func CreatePost(c *fiber.Ctx) error {
	var blogpost models.Post
	if err := c.BodyParser(&blogpost); err != nil {
		fmt.Println("Error parsing body")
	}
	if err := database.DB.Create(&blogpost).Error; err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Unable to create post",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Post created successfully",
	})
}

// AllPost returns all posts
func AllPost(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit := 5
	offset := (page - 1) * limit
	var total int64
	var posts []models.Post
	database.DB.Preload("User").Offset(offset).Limit(limit).Find(&posts)
	database.DB.Model(&models.Post{}).Count(&total)

	return c.JSON(fiber.Map{
		"data": posts,
		"meta": fiber.Map{
			"total":     total,
			"page":      page,
			"last_page": math.Ceil(float64(total / int64(limit))),
		},
	})
}

func DetailPost(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var post models.Post
	database.DB.Preload("User").Where("id = ?", id).First(&post)
	return c.JSON(fiber.Map{
		"data": post,
	})
}
