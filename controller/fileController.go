package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"math/rand"
	"mime"
	"os"
	"path/filepath"
)

const MaxUploadSize = 10 * 1024 * 1024 // 10 MB
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func SaveFile(c *fiber.Ctx, dirPath string) (string, error) {
	// Define a list of allowed MIME types
	allowedMimeTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"text/plain":      true, // Allow text files
		"application/pdf": true, // Allow PDF files
		// Add more allowed MIME types here
	}

	// Parse the uploaded file from the request
	file, err := c.FormFile("file")
	if err != nil {
		return "", err
	}

	// Check the file's MIME type to ensure it's a safe file format
	ext := filepath.Ext(file.Filename)
	mimeType := mime.TypeByExtension(ext)
	if !allowedMimeTypes[mimeType] {
		return "", fiber.NewError(fiber.StatusBadRequest, "unsupported file format")
	}

	// Check the file's size to ensure it's within the allowed limit
	if file.Size > MaxUploadSize {
		return "", fiber.NewError(fiber.StatusBadRequest, "file size exceeds the limit")
	}

	// Save the file in the directory corresponding to the post ID
	filename := RandomString(8) + "-" + file.Filename
	filePath := dirPath + "/" + filename
	if err := c.SaveFile(file, filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

func UploadFile(c *fiber.Ctx) error {
	// Extract the post ID from the request
	postID := c.Params("postID")

	// Check if a directory with the post ID exists. If not, create it.
	dirPath := "./uploads/" + postID
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dirPath, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}

	filePath, err := SaveFile(c, dirPath)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"url": c.BaseURL() + "/api/uploads/" + postID + "/" + filepath.Base(filePath),
	})
}
