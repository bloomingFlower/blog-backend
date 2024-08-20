package controller

import (
	"strconv"
	"time"

	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

// GetComments get comments
func GetComments(c *fiber.Ctx) error {
	postID := c.Params("postId")
	var comments []models.Comment

	result := database.DB.Preload("User").Preload("Votes").Preload("Children").Where("post_id = ? AND parent_id IS NULL", postID).Find(&comments)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Error("--> CommentController: GetComments: Failed to get comments: ", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get comments",
		})
	}

	return c.JSON(fiber.Map{
		"data": comments,
	})
}

// CreateComment create comment
func CreateComment(c *fiber.Ctx) error {
	log.Debug("--> CommentController: CreateComment: c: ", c)

	postID := c.Params("postId")
	if postID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Post ID is required",
		})
	}

	var comment models.Comment
	if err := c.BodyParser(&comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	// Set PostID from URL parameter
	postIDUint, err := strconv.ParseUint(postID, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Post ID",
		})
	}
	comment.PostID = uint(postIDUint)

	// Get IP address
	ip := c.IP()

	// Apply limit if not 127.0.0.1
	if ip != "127.0.0.1" {
		// Check the number of comments written within the last minute from the same IP
		var count int64
		if err := database.DB.Model(&models.Comment{}).Where("ip_address = ? AND created_at > ?", ip, time.Now().Add(-1*time.Minute)).Count(&count).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check comment count",
			})
		}

		if count >= 5 {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Comment writing limit exceeded. Please try again later.",
			})
		}
	}

	// Set userID if provided
	log.Debug("--> CommentController: CreateComment: comment: ", comment)
	log.Debug("--> CommentController: CreateComment: postID: ", comment.PostID)
	log.Debug("--> CommentController: CreateComment: content: ", comment.Content)
	log.Debug("--> CommentController: CreateComment: comment.UserID: ", comment.UserID)
	if comment.UserID != nil && *comment.UserID != 0 {
		// Optionally, you can verify if the user exists in the database
		var user models.User
		if err := database.DB.First(&user, comment.UserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid User ID",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify user",
			})
		}
	} else {
		comment.UserID = nil // Ensure it's nil for anonymous comments
	}

	comment.IPAddress = ip
	log.Debug("--> CommentController: CreateComment: comment: ", comment)
	if err := database.DB.Create(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create comment",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Comment created successfully",
		"data":    comment,
	})
}

// CreateReply create reply to comment
func CreateReply(c *fiber.Ctx) error {
	parentIDStr := c.Params("commentId")
	parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid parent comment ID",
		})
	}

	var reply models.Comment
	if err := c.BodyParser(&reply); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	// Get IP address
	ip := c.IP()

	// Apply limit if not 127.0.0.1
	if ip != "127.0.0.1" {
		// Check the number of comments written within the last minute from the same IP
		var count int64
		if err := database.DB.Model(&models.Comment{}).Where("ip_address = ? AND created_at > ?", ip, time.Now().Add(-1*time.Minute)).Count(&count).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check comment count",
			})
		}

		if count >= 5 {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Comment writing limit exceeded. Please try again later.",
			})
		}
	}

	// Set ParentID
	if err := database.DB.First(&models.Comment{}, parentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Parent comment not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check parent comment",
		})
	}

	parentIDUint := uint(parentID)
	reply.ParentID = &parentIDUint
	reply.IPAddress = ip

	if err := database.DB.Create(&reply).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create reply",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Reply created successfully",
		"data":    reply,
	})
}

// VoteComment vote comment
func VoteComment(c *fiber.Ctx) error {
	commentIDStr := c.Params("commentId")
	commentID, err := strconv.ParseUint(commentIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	var vote models.Vote
	if err := c.BodyParser(&vote); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	// Get IP address
	ip := c.IP()

	// Check if the comment exists
	if err := database.DB.First(&models.Comment{}, commentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Comment not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check comment",
		})
	}

	vote.CommentID = uint(commentID)
	vote.IPAddress = ip

	// Check if the vote already exists, and if it does, update it
	var existingVote models.Vote
	err = database.DB.Where("comment_id = ? AND ip_address = ?", vote.CommentID, ip).First(&existingVote).Error
	if err == nil {
		// Update the existing vote
		existingVote.Emoji = vote.Emoji
		if err := database.DB.Save(&existingVote).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update vote",
			})
		}
	} else if err == gorm.ErrRecordNotFound {
		// Create a new vote
		if err := database.DB.Create(&vote).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create vote",
			})
		}
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process vote",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Vote processed successfully",
	})
}
