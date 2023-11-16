package controller

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"math/rand"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func UploadImage(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["image"]
	filename := ""
	for _, file := range files {
		filename = RandomString(6) + "-" + file.Filename
		if err := c.SaveFile(file, "./uploads/img/"+filename); err != nil {
			return err
		}
	}
	log.Println(filename)
	return c.JSON(fiber.Map{
		"url": "http://localhost:3000/api/uploads/img/" + filename,
	})

}
