package controller

import (
	"strconv"

	"github.com/bloomingFlower/blog-backend/database"
	"github.com/bloomingFlower/blog-backend/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GetAboutInfo returns the about me information for a specific ID from the database
func GetAboutInfo(c *fiber.Ctx) error {
	// Get AboutInfo ID from the URL parameter
	id := c.Params("id")

	var aboutInfo models.AboutInfo

	// Fetch specific AboutInfo from the database
	if err := database.DB.Preload("Contacts").Preload("Sections.Items").First(&aboutInfo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "About information not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch about information",
		})
	}

	// Return the about info as JSON
	return c.JSON(fiber.Map{
		"data": aboutInfo,
	})
}

// CreateAboutInfo creates a new AboutInfo record in the database
func CreateAboutInfo(c *fiber.Ctx) error {
	// Parse request body
	var aboutInfo models.AboutInfo
	if err := c.BodyParser(&aboutInfo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	// Create record in the database
	if err := database.DB.Create(&aboutInfo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create about information",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "About information created successfully",
		"data":    aboutInfo,
	})
}

// UpdateAboutInfo updates an existing AboutInfo record in the database
func UpdateAboutInfo(c *fiber.Ctx) error {
	id := c.Params("id")

	// Find existing record
	var aboutInfo models.AboutInfo
	if err := database.DB.First(&aboutInfo, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "About information not found",
		})
	}

	// Parse request body
	if err := c.BodyParser(&aboutInfo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	// Update record in the database
	if err := database.DB.Save(&aboutInfo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update about information",
		})
	}

	return c.JSON(fiber.Map{
		"message": "About information updated successfully",
		"data":    aboutInfo,
	})
}

// DeleteAboutInfo deletes an existing AboutInfo record from the database
func DeleteAboutInfo(c *fiber.Ctx) error {
	id := c.Params("id")

	// Delete record from the database
	if err := database.DB.Delete(&models.AboutInfo{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete about information",
		})
	}

	return c.JSON(fiber.Map{
		"message": "About information deleted successfully",
	})
}

// CreateSection creates a new Section for a specific AboutInfo
func CreateSection(c *fiber.Ctx) error {
	// Get AboutInfo ID from the URL parameter
	aboutInfoID := c.Params("aboutInfoID")

	// Parse request body
	var section models.Section
	if err := c.BodyParser(&section); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	// Convert aboutInfoID to uint
	aboutInfoIDUint, err := strconv.ParseUint(aboutInfoID, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid AboutInfo ID",
		})
	}
	// Set the AboutInfoID
	section.AboutInfoID = uint(aboutInfoIDUint)

	// Create record in the database
	if err := database.DB.Create(&section).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create section",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Section created successfully",
		"data":    section,
	})
}

// UpdateSection updates an existing Section
func UpdateSection(c *fiber.Ctx) error {
	aboutInfoID := c.Params("aboutInfoID")
	sectionID := c.Params("sectionID")

	var section models.Section
	if err := database.DB.Where("id = ? AND about_info_id = ?", sectionID, aboutInfoID).First(&section).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Section not found",
		})
	}

	if err := c.BodyParser(&section); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	if err := database.DB.Save(&section).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update section",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Section updated successfully",
		"data":    section,
	})
}

// DeleteSection deletes an existing Section
func DeleteSection(c *fiber.Ctx) error {
	aboutInfoID := c.Params("aboutInfoID")
	sectionID := c.Params("sectionID")

	if err := database.DB.Where("id = ? AND about_info_id = ?", sectionID, aboutInfoID).Delete(&models.Section{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete section",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Section deleted successfully",
	})
}

// CreateSectionItem creates a new SectionItem for a specific Section
func CreateSectionItem(c *fiber.Ctx) error {
	aboutInfoID := c.Params("aboutInfoID")
	sectionID := c.Params("sectionID")

	// Parse request body
	var sectionItem models.SectionItem
	if err := c.BodyParser(&sectionItem); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	// Convert sectionID to uint
	sectionIDUint, err := strconv.ParseUint(sectionID, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Section ID",
		})
	}
	// Set the SectionID
	sectionItem.SectionID = uint(sectionIDUint)

	// Verify that the section belongs to the specified AboutInfo
	var section models.Section
	if err := database.DB.Where("id = ? AND about_info_id = ?", sectionID, aboutInfoID).First(&section).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Section not found for the specified AboutInfo",
		})
	}

	// Create record in the database
	if err := database.DB.Create(&sectionItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create section item",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Section item created successfully",
		"data":    sectionItem,
	})
}

// UpdateSectionItem updates an existing SectionItem
func UpdateSectionItem(c *fiber.Ctx) error {
	aboutInfoID := c.Params("aboutInfoID")
	sectionID := c.Params("sectionID")
	itemID := c.Params("itemID")

	var sectionItem models.SectionItem
	if err := database.DB.Joins("JOIN sections ON section_items.section_id = sections.id").
		Where("section_items.id = ? AND sections.id = ? AND sections.about_info_id = ?", itemID, sectionID, aboutInfoID).
		First(&sectionItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Section item not found",
		})
	}

	// Parse request body
	var updateData map[string]interface{}
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	// Remove created_at, updated_at, and id from updateData if present
	delete(updateData, "created_at")
	delete(updateData, "updated_at")
	delete(updateData, "id")

	// Update only the fields provided in the request
	if err := database.DB.Model(&sectionItem).Omit("CreatedAt", "UpdatedAt").Updates(updateData).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to update section item",
			"message": err.Error(),
		})
	}

	// Fetch the updated item
	if err := database.DB.First(&sectionItem, sectionItem.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch updated section item",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Section item updated successfully",
		"data":    sectionItem,
	})
}

// DeleteSectionItem deletes an existing SectionItem
func DeleteSectionItem(c *fiber.Ctx) error {
	aboutInfoID := c.Params("aboutInfoID")
	sectionID := c.Params("sectionID")
	itemID := c.Params("itemID")

	if err := database.DB.Joins("JOIN sections ON section_items.section_id = sections.id").
		Where("section_items.id = ? AND sections.id = ? AND sections.about_info_id = ?", itemID, sectionID, aboutInfoID).
		Delete(&models.SectionItem{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete section item",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Section item deleted successfully",
	})
}
