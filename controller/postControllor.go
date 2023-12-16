package controller

import (
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
	// JSON 형식의 해시태그를 Go 슬라이스로 변환
	tags := strings.Split(tagsJSON, ",") // ["fdg", "hgfj", "dsfg", "gfhj"]	err = json.Unmarshal([]byte(tagsJSON), &tags)

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
	if _, err := c.FormFile("file"); err == nil {
		filePath, err = SaveFile(c, dirPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Could not save file",
			})
		}
		blogpost.File = filePath
		if err = database.DB.Save(&blogpost).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Unable to save file path",
			})
		}
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

	// Check if the user is logged in
	cookie := c.Cookies("jwt")
	idStr, err := util.ParseJwt(cookie)
	if err != nil {
		idStr = "0"
	}
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	userId := uint(idInt)
	if userId == 0 {
		// If not logged in, exclude hidden posts
		database.DB.Preload("User").Where("hidden = ? OR hidden IS NULL", false).Offset(offset).Limit(limit).Find(&posts)
		database.DB.Model(&models.Post{}).Where("hidden = ? OR hidden IS NULL", false).Count(&total)
	} else {
		database.DB.Preload("User").Offset(offset).Limit(limit).Find(&posts)
		database.DB.Model(&models.Post{}).Count(&total)
	}

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
	// Extract the post ID from the request
	postID := c.Params("id")
	// Check if postID is a valid integer
	if _, err := strconv.Atoi(postID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "postID must be a valid integer",
		})
	}

	// Find the post with the given ID
	var post models.Post
	// Check if the user is logged in
	cookie := c.Cookies("jwt")
	idStr, err := util.ParseJwt(cookie)
	if err != nil {
		idStr = "0"
	}
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	userId := uint(idInt)
	if userId == 0 {
		// If not logged in, exclude hidden posts
		result := database.DB.Preload("User").Where("id = ? AND (hidden = ? OR hidden IS NULL)", postID, false).First(&post)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Post not found",
			})
		}
	} else {
		result := database.DB.Preload("User").Where("id = ?", postID).First(&post)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Post not found",
			})
		}
	}

	// Return the post details
	return c.JSON(fiber.Map{
		"data": post,
	})
}

func UpdatePost(c *fiber.Ctx) error {
	// Extract the post ID from the request
	postID := c.Params("id")

	// Check if postID is a valid integer
	if _, err := strconv.Atoi(postID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "postID must be a valid integer",
		})
	}

	// 제목, Quill의 내용, 해시태그 파싱
	// Check if the user is logged in
	cookie := c.Cookies("jwt")
	idStr, err := util.ParseJwt(cookie)
	if err != nil {
		idStr = "0"
	}
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	userID := uint(idInt)

	title := c.FormValue("title")
	content := c.FormValue("content")
	tagsJSON := c.FormValue("tags") // 해시태그는 JSON 형식의 문자열로 가정
	log.Println("c.FormValue(content): ", c.FormValue("content"))
	log.Println("c.FormValue(tags): ", c.FormValue("tags"))
	log.Println("c.Locals(userID): ", c.Locals("userID"))
	log.Println("c.FormValue(title): ", c.FormValue("title"))
	// JSON 형식의 해시태그를 Go 슬라이스로 변환
	// Convert the tags slice back to JSON
	tags := strings.Split(tagsJSON, ",") // ["fdg", "hgfj", "dsfg", "gfhj"]	err = json.Unmarshal([]byte(tagsJSON), &tags)
	if err != nil {
		log.Println("Error parsing tags:", err)
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
	if err := c.BodyParser(&blogpost); err != nil {
		fmt.Println("Error parsing body")
	}
	result := database.DB.Model(&blogpost).Where("id = ?", postID).Updates(blogpost)
	if result.Error != nil {
		log.Println("Error updating post:", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error updating post",
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
	if _, err := c.FormFile("file"); err == nil {
		filePath, err = SaveFile(c, dirPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Could not save file",
			})
		}
		blogpost.File = filePath
		log.Println("filePath: ", filePath)
		if err = database.DB.Save(&blogpost).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Unable to save file path",
			})
		}
	}

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

func SearchPost(c *fiber.Ctx) error {
	// Get the search query, type, and pagination parameters from the request
	query := c.Query("query", "")
	searchType := c.Query("type", "all")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit := 5
	offset := (page - 1) * limit

	// Initialize an empty slice to hold the results
	var posts []models.Post
	var total int64

	// Perform the search based on the type
	switch searchType {
	case "title":
		database.DB.Where("title LIKE ?", "%"+query+"%").Offset(offset).Limit(limit).Find(&posts)
		database.DB.Model(&models.Post{}).Where("title LIKE ?", "%"+query+"%").Count(&total)
	case "content":
		database.DB.Where("content LIKE ?", "%"+query+"%").Offset(offset).Limit(limit).Find(&posts)
		database.DB.Model(&models.Post{}).Where("content LIKE ?", "%"+query+"%").Count(&total)
	case "tags":
		database.DB.Where("tags LIKE ?", "%"+query+"%").Offset(offset).Limit(limit).Find(&posts)
		database.DB.Model(&models.Post{}).Where("tags LIKE ?", "%"+query+"%").Count(&total)
	default: // Search all fields
		database.DB.Where("title LIKE ? OR content LIKE ? OR tags LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%").Offset(offset).Limit(limit).Find(&posts)
		database.DB.Model(&models.Post{}).Where("title LIKE ? OR content LIKE ? OR tags LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%").Count(&total)
	}

	// Calculate the last page number
	lastPage := int(math.Ceil(float64(total) / float64(limit)))

	// Return the search results and pagination info
	return c.JSON(fiber.Map{
		"data": posts,
		"meta": fiber.Map{
			"total":     total,
			"page":      page,
			"last_page": lastPage,
		},
	})
}

func HidePost(c *fiber.Ctx) error {
	// Extract the post ID from the request
	postID := c.Params("id")

	// Find the post with the given ID
	var post models.Post
	result := database.DB.Where("id = ?", postID).First(&post)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Post not found",
		})
	}

	// Hide the post
	if post.Hidden {
		post.Hidden = false
	} else {
		post.Hidden = true
	}

	// Save the current state of the post.Hidden field
	hidden := post.Hidden
	// 	result = database.DB.Model(&post).Updates(models.Post{Hidden: hidden}) 이건 zero value 무시
	result = database.DB.Model(&post).UpdateColumn("hidden", hidden)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error updating post",
		})
	}

	// Return a success message
	return c.JSON(fiber.Map{
		"message": "Post hidden successfully",
	})
}
