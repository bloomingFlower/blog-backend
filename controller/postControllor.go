package controller

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"html"
	"regexp"

	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/bloomingFlower/blog-backend/util"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gorilla/feeds"
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
	tags := []string{}
	// JSON 형식의 해시태그를 Go 슬라이스로 변환
	if tagsJSON != "" {
		tags = strings.Split(tagsJSON, ",") // ["fdg", "hgfj", "dsfg", "gfhj"]	err = json.Unmarshal([]byte(tagsJSON), &tags)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Error parsing tags",
			})
		}
	}

	category := c.FormValue("category")

	// 데이터베이스에 저장
	blogpost := models.Post{
		UserID:    userID,
		Title:     title,
		Content:   content,
		Tags:      strings.Join(tags, ","),
		Category:  category,
		UpdatedAt: nil,
	}
	if err := database.DB.Create(&blogpost).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unable to create post",
		})
	}

	// 포스트 ID를 기준으로 디렉토리 생성
	dirPath := fmt.Sprintf("uploads/%d", blogpost.ID)
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
		blogpost.File = fmt.Sprintf("files/%d/%s", blogpost.ID, filePath)
		log.Debug("--> PostController: CreatePost: blogpost.File: ", blogpost.File)
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
	limit := 6
	offset := (page - 1) * limit
	var total int64
	var posts []models.Post

	category := c.Query("category", "")

	// Check if the user is logged in
	cookie := c.Cookies("jwt")
	idStr, err := util.ParseJwt(cookie)
	if err != nil {
		idStr = "0"
	}
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	userId := uint(idInt)

	// Generate the base query
	query := database.DB.Preload("User").Order("created_at DESC")

	// Apply category filter
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if userId == 0 {
		// If not logged in, exclude hidden posts
		query = query.Where("hidden = ? OR hidden IS NULL", false)
	}

	// Apply pagination and retrieve results
	query.Offset(offset).Limit(limit).Find(&posts)

	// Count the total number of posts
	countQuery := database.DB.Model(&models.Post{})
	if category != "" {
		countQuery = countQuery.Where("category = ?", category)
	}
	if userId == 0 {
		countQuery = countQuery.Where("hidden = ? OR hidden IS NULL", false)
	}
	countQuery.Count(&total)

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
		log.Error(err.Error())
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

	// Check if the user is logged in
	cookie := c.Cookies("jwt")
	idStr, err := util.ParseJwt(cookie)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	// Find the post
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Post not found",
		})
	}

	// Check if the user is the owner of the post
	if post.UserID != uint(userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You don't have permission to update this post",
		})
	}

	// 제목, Quill의 내용, 해시태그 파싱
	title := c.FormValue("title")
	content := c.FormValue("content")
	tagsJSON := c.FormValue("tags") // 해시태그는 JSON 형식의 문자열로 가정
	tags := []string{}
	if tagsJSON != "" {
		tags = strings.Split(tagsJSON, ",") // ["fdg", "hgfj", "dsfg", "gfhj"]	err = json.Unmarshal([]byte(tagsJSON), &tags)
		if err != nil {
			log.Error("Error parsing tags:", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Error parsing tags",
			})
		}
	}

	category := c.FormValue("category")

	// 현재 시간을 UpdatedAt으로 설정
	now := time.Now()
	blogpost := models.Post{
		UserID:    uint(userID),
		Title:     title,
		Content:   content,
		Tags:      strings.Join(tags, ","),
		Category:  category,
		UpdatedAt: &now,
	}
	if err := c.BodyParser(&blogpost); err != nil {
		fmt.Println("Error parsing body")
	}
	result := database.DB.Model(&blogpost).Where("id = ?", postID).Updates(blogpost)
	if result.Error != nil {
		log.Error("Error updating post:", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error updating post",
		})
	}
	// 포스트 ID를 기준으로 디렉토리 생성
	dirPath := fmt.Sprintf("uploads/%d", post.ID)
	_, err = os.Stat(dirPath)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dirPath, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}

	// 파일이 있을 경우, 파일 저장
	filename := ""
	if _, err := c.FormFile("file"); err == nil {
		// 기존 파일 삭제
		if post.File != "" {
			log.Debug("--> PostController: UpdatePost: post.ID: ", post.ID)
			log.Debug("--> PostController: UpdatePost: post.File: ", post.File)
			oldFilePath := filepath.Join("uploads", strconv.Itoa(int(post.ID)), filepath.Base(post.File))
			if err := os.Remove(oldFilePath); err != nil {
				log.Error("Error deleting old file: %v", err)
			}
		}

		// 새 파일 저장
		filename, err = SaveFile(c, dirPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Could not save file",
			})
		}
		log.Debug("--> PostController: UpdatePost: post.ID: ", post.ID)
		post.File = fmt.Sprintf("files/%d/%s", post.ID, filename)
		log.Debug("--> PostController: UpdatePost: post.File: ", post.File)
		if err = database.DB.Model(&post).Update("file", post.File).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Unable to update file path",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Post updated successfully",
	})
}

func ServeFile(c *fiber.Ctx) error {
	id := c.Params("id")
	filename := c.Params("filename")
	path := fmt.Sprintf("uploads/%s/%s", id, filename)
	log.Debug("--> PostController: ServeFile: path: ", path)

	// Content-Disposition 헤더 설정
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	c.Type(filepath.Ext(filename))

	return c.SendFile(path)
}

func UniquePost(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")
	id, err := util.ParseJwt(cookie)
	if err != nil {
		log.Error(err)
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

	// Check if the user is logged in
	cookie := c.Cookies("jwt")
	idStr, err := util.ParseJwt(cookie)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	// Find the post
	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Post not found",
		})
	}

	// Check if the user is the owner of the post
	if post.UserID != uint(userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You don't have permission to delete this post",
		})
	}

	// Delete the post
	deleteQuery := database.DB.Delete(&post)
	if deleteQuery.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Unable to delete post",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Post deleted successfully",
	})
}

// TODO : ElasticSearch 적용
func SearchPost(c *fiber.Ctx) error {
	// Get the search query, type, and pagination parameters from the request
	query := c.Query("query", "")
	searchType := c.Query("type", "all")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit := 6
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

// RSSFeed generates an RSS feed of recent posts
func RSSFeed(c *fiber.Ctx) error {
	log.Debug("--> PostController: RSSFeed: ")
	var posts []models.Post
	limit := 10 // Get the latest 10 posts

	result := database.DB.Where("hidden = ? OR hidden IS NULL", false).
		Order("created_at DESC").
		Limit(limit).
		Find(&posts)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error fetching posts",
		})
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8008"
	}

	feed := &feeds.Feed{
		Title:       "Our Journey",
		Link:        &feeds.Link{Href: baseURL},
		Description: "Recent posts from Our Journey",
		Author:      &feeds.Author{Name: "Our Journey"},
		Created:     time.Now(),
	}

	for _, post := range posts {
		cleanContent := cleanHTMLContent(post.Content)

		item := &feeds.Item{
			Title:       post.Title,
			Link:        &feeds.Link{Href: fmt.Sprintf("%s/post/%d", baseURL, post.ID)},
			Description: cleanContent,
			Author:      &feeds.Author{Name: post.User.FirstName + " " + post.User.LastName},
			Created:     post.CreatedAt,
		}
		feed.Items = append(feed.Items, item)
	}

	rss, err := feed.ToRss()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error generating RSS feed",
		})
	}

	rss = strings.Replace(rss, `<?xml version="1.0"?>`, `<?xml version="1.0" encoding="UTF-8"?>`, 1)

	c.Set("Content-Type", "application/rss+xml; charset=utf-8")
	return c.Send([]byte(rss))
}

func cleanHTMLContent(content string) string {
	// Remove HTML tags
	re := regexp.MustCompile("<[^>]*>")
	cleanContent := re.ReplaceAllString(content, "")

	// Decode HTML entities
	cleanContent = html.UnescapeString(cleanContent)

	return cleanContent
}
