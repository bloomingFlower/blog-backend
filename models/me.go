package models

import "gorm.io/gorm"

type AboutInfo struct {
	gorm.Model
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Contacts    []Contact `json:"contacts" gorm:"foreignKey:AboutInfoID"`
	Sections    []Section `json:"sections" gorm:"foreignKey:AboutInfoID"`
}

type Contact struct {
	gorm.Model
	AboutInfoID uint   `json:"-"`
	Icon        string `json:"icon"`
	Label       string `json:"label"`
	Link        string `json:"link"`
}

type Section struct {
	gorm.Model
	AboutInfoID uint          `json:"-"`
	Title       string        `json:"title"`
	Icon        string        `json:"icon"`
	Items       []SectionItem `json:"items" gorm:"foreignKey:SectionID"`
}

type SectionItem struct {
	gorm.Model
	SectionID   uint   `json:"-"`
	Title       string `json:"title"`
	Institution string `json:"institution,omitempty"`
	Company     string `json:"company,omitempty"`
	Year        string `json:"year,omitempty"`
	Period      string `json:"period,omitempty"`
	Description string `json:"description,omitempty"`
}
