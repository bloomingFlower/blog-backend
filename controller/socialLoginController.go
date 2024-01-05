package controller

import (
	"context"
	"encoding/json"
	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"io"
	"net/http"
	"os"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_PD"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	githubOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/github/callback",
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_PD"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
)

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
	url := githubOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return c.Redirect(url)
}

func GithubCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	token, err := githubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(http.StatusUnauthorized).SendString("Unauthorized")
	}

	client := githubOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Internal server error")
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
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
			FirstName: userData["name"].(string),
			Email:     userData["email"].(string),
			Picture:   userData["avatar_url"].(string),
		}
		database.DB.Create(&user)
	}

	return c.JSON(fiber.Map{
		"user":    user,
		"message": "User logged in successfully",
	})
}
