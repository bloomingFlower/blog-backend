package controller

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bloomingFlower/blog-backend/util"
	"github.com/dgrijalva/jwt-go"

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
}

func Login(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println("unable to parse body")
	}
	var user models.User
	database.DB.Where("email = ?", data["email"]).First(&user)
	if user.ID == 0 {
		c.Status(http.StatusNotFound)
		return c.JSON(fiber.Map{
			"message": "User not found",
		})
	}
	if err := user.ComparePassword(data["password"]); err != nil {
		c.Status(http.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Incorrect password",
		})
	}
	token, err := util.GenerateJwt(user.ID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"message": "Unable to login",
		})
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)
	c.Status(http.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Successfully login",
		"user":    user,
	})
}

func DeleteUser(c *fiber.Ctx) error {
	// Get the JWT token from the request cookies
	cookie := c.Cookies("jwt")

	// Parse the JWT token to get the user ID
	id, err := util.ParseJwt(cookie)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// Find and delete the user with the given ID
	var user models.User
	database.DB.Where("id = ?", id).First(&user)
	if user.ID == 0 {
		c.Status(http.StatusNotFound)
		return c.JSON(fiber.Map{
			"message": "User not found",
		})
	}
	database.DB.Delete(&user)

	c.Status(http.StatusOK)
	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

type Claims struct {
	jwt.StandardClaims
}
