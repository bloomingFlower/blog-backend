package controller

import (
	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/gofiber/fiber/v2"
	"time"
)

func SaveAPILog(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		// Handle the error appropriately, for example:
		userID = 0
	}
	// Record the time when the request starts
	startTime := time.Now()

	// Call the next handler in the stack and wait for it to complete
	err := c.Next()
	if err != nil {
		return err
	}

	// Record the time when the request ends
	endTime := time.Now()

	// Calculate the response time
	responseTime := endTime.Sub(startTime)

	method := c.Method()
	url := c.OriginalURL()
	ip := c.IP()
	statusCode := c.Response().StatusCode()

	// Create a new APILog instance
	log := models.APILog{
		RequestMethod:      method,
		RequestURL:         url,
		RequestIP:          ip,
		ResponseStatusCode: statusCode,
		ResponseTime:       responseTime,
		UserID:             userID,
	}

	// Save the APILog instance to the database
	if err := database.DB.Create(&log).Error; err != nil {
		return err
	}

	return nil
}
