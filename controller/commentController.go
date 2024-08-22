package controller

import (
	"strconv"
	"time"

	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/microcosm-cc/bluemonday"
	"gorm.io/gorm"
)

// GetComments get comments
func GetComments(c *fiber.Ctx) error {
	postID := c.Params("postId")
	var comments []models.Comment

	// Preload User and Children's User, selecting specific fields
	result := database.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID", "FirstName", "Picture")
	}).Preload("Votes").Preload("Children", func(db *gorm.DB) *gorm.DB {
		return db.Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID", "FirstName", "Picture")
		}).Preload("Votes")
	}).Where("post_id = ? AND parent_id IS NULL", postID).Find(&comments)

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

	// Sanitize comment content
	p := bluemonday.UGCPolicy()
	comment.Content = p.Sanitize(comment.Content)

	// Check comment length after sanitization
	if len(comment.Content) > 3000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Comment content exceeds 3000 characters limit",
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
			"data":  reply,
		})
	}

	// Sanitize reply content
	p := bluemonday.UGCPolicy()
	reply.Content = p.Sanitize(reply.Content)

	// Check reply length after sanitization
	if len(reply.Content) > 3000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Reply content exceeds 3000 characters limit",
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

	// Find parent comment to get post_id
	var parentComment models.Comment
	if err := database.DB.First(&parentComment, parentID).Error; err != nil {
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
	reply.PostID = parentComment.PostID // Set PostID from parent comment
	reply.IPAddress = ip
	// Set userID if provided
	if reply.UserID != nil && *reply.UserID != 0 {
		// Optionally, you can verify if the user exists in the database
		var user models.User
		if err := database.DB.First(&user, reply.UserID).Error; err != nil {
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
		reply.UserID = nil // Ensure it's nil for anonymous comments
	}

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

	// Sanitize emoji
	p := bluemonday.UGCPolicy()
	vote.Emoji = p.Sanitize(vote.Emoji)

	// Check emoji length
	if len(vote.Emoji) > 20 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Emoji exceeds 20 characters limit",
		})
	}

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

	// Get IP address
	ip := c.IP()

	// Determine if we're using UserID or IP
	var query string
	var args []interface{}
	if vote.UserID != nil && *vote.UserID != 0 {
		// User is logged in, use UserID
		query = "comment_id = ? AND user_id = ?"
		args = []interface{}{vote.CommentID, *vote.UserID}
		vote.IPAddress = "" // Clear IP address for logged-in users
	} else {
		// Anonymous user, use IP
		query = "comment_id = ? AND ip_address = ?"
		args = []interface{}{vote.CommentID, ip}
		vote.UserID = nil // Clear UserID for anonymous users
		vote.IPAddress = ip
	}

	// Check if the vote already exists
	var existingVote models.Vote
	err = database.DB.Where(query, args...).First(&existingVote).Error
	if err == nil {
		// Update the existing vote
		existingVote.Emoji = vote.Emoji
		if err := database.DB.Save(&existingVote).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update vote",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Vote updated successfully",
			"data":    existingVote,
		})
	} else if err == gorm.ErrRecordNotFound {
		// Create a new vote
		if err := database.DB.Create(&vote).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create vote",
			})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Vote created successfully",
			"data":    vote,
		})
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process vote",
		})
	}
}

// DeleteComment deletes a comment
func DeleteComment(c *fiber.Ctx) error {
	// Get comment ID from URL parameter
	commentIDStr := c.Params("commentId")
	commentID, err := strconv.ParseUint(commentIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	// Find the comment
	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Comment not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to find comment",
		})
	}

	// Check if the user is authorized to delete the comment
	// This is a placeholder. You should implement proper authorization logic.
	// For example, check if the user is an admin or the comment owner.
	isAuthorized := true // Replace with actual authorization logic

	if !isAuthorized {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not authorized to delete this comment",
		})
	}

	// Soft delete the comment
	if err := database.DB.Model(&comment).Update("deleted_at", time.Now()).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete comment",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Comment deleted successfully",
	})
}
