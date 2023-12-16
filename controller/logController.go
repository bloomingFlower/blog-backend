package controller

import (
	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/bloomingFlower/blog-backend/util"
	"github.com/gofiber/fiber/v2"
	"log"
	"net/http"
	"strconv"
	"time"
)

func SaveAPILog(c *fiber.Ctx) error {
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
	// Record the time when the request starts
	startTime := time.Now()

	// Record the time when the request ends
	endTime := time.Now()

	// Calculate the response time
	responseTime := endTime.Sub(startTime)

	method := c.Method()
	url := c.OriginalURL()
	ip := c.IP()
	statusCode := c.Response().StatusCode()

	// Create a new APILog instance
	logData := models.APILog{
		RequestMethod:      method,
		RequestURL:         url,
		RequestIP:          ip,
		ResponseStatusCode: statusCode,
		ResponseTime:       responseTime,
		UserID:             userId,
	}

	// Save the APILog instance to the database
	if err := database.DB.Create(&logData).Error; err != nil {
		log.Println(err.Error())
		return err
	}

	c.Status(http.StatusOK)
	return c.JSON(fiber.Map{
		"message": "APILog saved successfully",
	})
}
