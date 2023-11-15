package controller

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"math"
	"strconv"

	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/bloomingFlower/blog-backend/util"
	"github.com/gofiber/fiber/v2"
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

	lastPage := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(fiber.Map{
		"data": posts,
		"meta": fiber.Map{
			"total":     total,
			"page":      page,
			"last_page": lastPage,
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

func UpdatePost(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var post models.Post
	if err := c.BodyParser(&post); err != nil {
		fmt.Println("Error parsing body")
	}
	database.DB.Model(&post).Where("id = ?", id).Updates(post)
	return c.JSON(fiber.Map{
		"message": "Post updated successfully",
	})
}

func UniquePost(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")
	id, err := util.ParseJwt(cookie)
	if err != nil {
		log.Println(err)
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	log.Println(cookie)
	var posts []models.Post
	database.DB.Model(&posts).Where("user_id=?", id).Preload("User").Find(&posts)
	database.DB.Debug().Model(&models.Post{}).Where("user_id=?", id).Preload("User").Find(&posts)
	return c.JSON(posts)
}

func DeletePost(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var post models.Post
	deleteQuery := database.DB.Where("id = ?", id).Delete(&post)
	if errors.Is(deleteQuery.Error, gorm.ErrRecordNotFound) {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Unable to delete post",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Post deleted successfully",
	})
}
