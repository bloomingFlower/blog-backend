package controller

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/gofiber/fiber/v2"
)

func validateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return re.MatchString(email)
}

func Register(c *fiber.Ctx) error {
	var data map[string]interface{}
	var userData models.User
	if err := c.BodyParser(&data); err != nil {
		fmt.Println("unable to parse body")
	}

	if len(data["password"].(string)) < 8 {
		c.Status(http.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Password must be at least 8 characters long",
		})
	}

	if !validateEmail(strings.TrimSpace(data["email"].(string))) {
		c.Status(http.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Invalid email address",
		})
	}

	database.DB.Where("email = ?", strings.TrimSpace(data["email"].(string))).First(&userData)
	if userData.ID != 0 {
		c.Status(http.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Email already exists",
		})
	}

	user := models.User{
		FirstName: data["first_name"].(string),
		LastName:  data["last_name"].(string),
		Email:     strings.TrimSpace(data["email"].(string)),
		Password:  []byte(data["password"].(string)),
		Phone:     data["phone"].(string),
	}

	user.SetPassword(data["password"].(string))
	err := database.DB.Create(&user)
	if err.Error != nil {
		log.Println(err.Error)
	}
	c.Status(http.StatusCreated)
	return c.JSON(fiber.Map{
		"user":    user,
		"message": "User created successfully",
	})

	//return c.SendString("Hello, World 👋!")
}
