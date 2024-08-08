package controller

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/bloomingFlower/blog-backend/util"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_PD"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
)

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Use FRONTEND_URL for Google redirect if not set
	if googleOauthConfig.RedirectURL == "" {
		googleOauthConfig.RedirectURL = os.Getenv("FRONTEND_URL") + "/auth/google/callback"
	}
}

func getGithubOauthConfig() *oauth2.Config {
	redirectURL := os.Getenv("FRONTEND_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080"
	}

	return &oauth2.Config{
		RedirectURL:  redirectURL + "/github-callback",
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_PD"),
		Scopes:       []string{"user", "user:email"},
		Endpoint:     github.Endpoint,
	}
}

func GoogleLogin(c *fiber.Ctx) error {
	url := googleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return c.Redirect(url)
}

func GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(http.StatusUnauthorized).SendString("Unauthorized")
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Internal server error")
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Internal server error")
	}
	var userData map[string]interface{}
	err = json.Unmarshal(data, &userData)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Internal server error")
	}

	var user models.User
	database.DB.Where("email = ?", userData["email"]).First(&user)
	if user.ID == 0 {
		user = models.User{
			FirstName: userData["given_name"].(string),
			LastName:  userData["family_name"].(string),
			Email:     userData["email"].(string),
			Picture:   userData["picture"].(string),
		}
		database.DB.Create(&user)
	}

	return c.JSON(fiber.Map{
		"user":    user,
		"message": "User logged in successfully",
	})
}

func GithubLogin(c *fiber.Ctx) error {
	githubOauthConfig := getGithubOauthConfig()
	url := githubOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("scope", "user:email"))
	return c.JSON(fiber.Map{
		"url": url,
	})
}

func GithubCallback(c *fiber.Ctx) error {
	githubOauthConfig := getGithubOauthConfig()
	code := c.Query("code")
	token, err := githubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Println("Token Exchange Error:", err)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	client := githubOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		log.Println("GitHub API Error:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Read Body Error:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	var userData map[string]interface{}
	err = json.Unmarshal(body, &userData)
	if err != nil {
		log.Println("JSON Unmarshal Error:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	githubID, ok := userData["id"].(float64)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Invalid GitHub ID",
		})
	}

	var user models.User
	result := database.DB.Where("github_id = ?", int64(githubID)).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// New user creation
			user = models.User{
				GithubID:  int64(githubID),
				FirstName: userData["login"].(string),
				Email:     userData["email"].(string),
			}
			if err := database.DB.Create(&user).Error; err != nil {
				log.Println("User Creation Error:", err)
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"message": "Failed to create user",
				})
			}
		} else {
			log.Println("Database Error:", result.Error)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error",
			})
		}
	}

	// JWT token generation
	jwtToken, err := util.GenerateJwt(user.ID)
	if err != nil {
		log.Println("JWT Generation Error:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Unable to login",
		})
	}

	// Set cookie
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    jwtToken,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "User logged in successfully",
		"user":    user,
		"token":   jwtToken,
	})
}
