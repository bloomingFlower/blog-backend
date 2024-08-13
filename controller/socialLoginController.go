package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
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
	googleOauthConfig *oauth2.Config
	githubOauthConfig *oauth2.Config
	frontendURL       string
)

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Set frontend URL
	frontendURL = os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:8080"
	}

	// Initialize Google OAuth config
	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_PD"),
		RedirectURL:  frontendURL + "/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Initialize GitHub OAuth config
	githubOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_PD"),
		RedirectURL:  frontendURL + "/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	// Log OAuth configurations
	log.Println("OAuth configurations initialized")
	log.Printf("Google RedirectURL: %s\n", googleOauthConfig.RedirectURL)
	log.Printf("GitHub RedirectURL: %s\n", githubOauthConfig.RedirectURL)
}

// generateStateOauthCookie generates a random state string and sets it in a cookie
func generateStateOauthCookie(c *fiber.Ctx) (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	cookie := fiber.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  time.Now().Add(time.Hour),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	return state, nil
}

// socialLogin is a generic function for handling social logins
func socialLogin(c *fiber.Ctx, oauthConfig *oauth2.Config) error {
	state, err := generateStateOauthCookie(c)
	if err != nil {
		log.Printf("Error generating state: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.JSON(fiber.Map{
		"url": url,
	})
}

func GoogleLogin(c *fiber.Ctx) error {
	return socialLogin(c, googleOauthConfig)
}

func GithubLogin(c *fiber.Ctx) error {
	return socialLogin(c, githubOauthConfig)
}

// socialCallback is a generic function for handling social login callbacks
func socialCallback(c *fiber.Ctx, oauthConfig *oauth2.Config, userInfoURL string, mapUserData func(map[string]interface{}) models.User) error {
	// State validation
	state := c.Query("state")
	oauthState := c.Cookies("oauthstate")
	if state != oauthState {
		log.Printf("Invalid oauth state, expected %s, got %s", oauthState, state)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid OAuth state",
		})
	}

	code := c.Query("code")
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Code exchange failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Code exchange failed",
		})
	}

	client := oauthConfig.Client(context.Background(), token)
	response, err := client.Get(userInfoURL)
	if err != nil {
		log.Printf("Failed getting user info: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed getting user info",
		})
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Failed reading response body: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed reading response body",
		})
	}

	var userData map[string]interface{}
	err = json.Unmarshal(data, &userData)
	if err != nil {
		log.Printf("Failed to parse user data: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse user data",
		})
	}

	user := mapUserData(userData)

	// Create or update user
	var existingUser models.User
	result := database.DB.Where("email = ?", user.Email).First(&existingUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			if err := database.DB.Create(&user).Error; err != nil {
				log.Printf("Failed to create user: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create user",
				})
			}
		} else {
			log.Printf("Database error: %v", result.Error)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Database error",
			})
		}
	} else {
		user.ID = existingUser.ID
		if err := database.DB.Save(&user).Error; err != nil {
			log.Printf("Failed to update user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update user",
			})
		}
	}

	// Generate JWT token
	jwtToken, err := util.GenerateJwt(user.ID)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT",
		})
	}

	// Set JWT token as cookie
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

func GoogleCallback(c *fiber.Ctx) error {
	return socialCallback(c, googleOauthConfig, "https://www.googleapis.com/oauth2/v2/userinfo", func(userData map[string]interface{}) models.User {
		return models.User{
			FirstName: userData["given_name"].(string),
			LastName:  userData["family_name"].(string),
			Email:     userData["email"].(string),
			Picture:   userData["picture"].(string),
		}
	})
}

func GithubCallback(c *fiber.Ctx) error {
	return socialCallback(c, githubOauthConfig, "https://api.github.com/user", func(userData map[string]interface{}) models.User {
		user := models.User{
			Email: userData["email"].(string),
		}
		if name, ok := userData["name"].(string); ok {
			names := strings.Split(name, " ")
			if len(names) > 0 {
				user.FirstName = names[0]
			}
			if len(names) > 1 {
				user.LastName = strings.Join(names[1:], " ")
			}
		}
		if avatar, ok := userData["avatar_url"].(string); ok {
			user.Picture = avatar
		}
		return user
	})
}
