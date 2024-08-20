package database

import (
	"log"
	"os"

	"github.com/bloomingFlower/blog-backend/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
	dsn := os.Getenv("DSN")
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	} else {
		log.Println("Database connected successfully")
	}
	DB = database
	err = database.AutoMigrate(&models.AboutInfo{}, &models.Contact{}, &models.Section{}, &models.SectionItem{}, &models.User{}, &models.Post{}, &models.APILog{}, &models.Comment{}, &models.Vote{})
	if err != nil {
		log.Fatal("Error migrating database: ", err)
	}
}
