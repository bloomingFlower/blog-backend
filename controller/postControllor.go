package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/bloomingFlower/blog-backend/util"
	"github.com/gofiber/fiber/v2"
)

func CreatePost(c *fiber.Ctx) error {
	// 제목, Quill의 내용, 해시태그 파싱
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	userID := uint(userID64)
	title := c.FormValue("title")
	content := c.FormValue("content")
	tagsJSON := c.FormValue("tags") // 해시태그는 JSON 형식의 문자열로 가정
	log.Println("c.FormValue(content): ", c.FormValue("content"))
	log.Println("c.FormValue(tags): ", c.FormValue("tags"))
	log.Println("c.Locals(userID): ", c.Locals("userID"))
	log.Println("c.FormValue(title): ", c.FormValue("title"))
	// JSON 형식의 해시태그를 Go 슬라이스로 변환
	var tags []string
	err = json.Unmarshal([]byte(tagsJSON), &tags)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing tags",
		})
	}

	// 데이터베이스에 저장
	blogpost := models.Post{
		UserID:  userID,
		Title:   title,
		Content: content,
		Tags:    strings.Join(tags, ","),
	}
	if err := database.DB.Create(&blogpost).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unable to create post",
		})
	}

	// 포스트 ID를 기준으로 디렉토리 생성
	dirPath := fmt.Sprintf("./uploads/%d/", blogpost.ID)
	_, err = os.Stat(dirPath)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dirPath, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}

	// 파일이 있을 경우, 파일 저장
	filePath := ""
	if _, err := c.FormFile("image"); err == nil {
		filePath, err = SaveFile(c, dirPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Could not save file",
			})
		}
		blogpost.Image = filePath
		log.Println("filePath: ", filePath)
		database.DB.Save(&blogpost)
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
