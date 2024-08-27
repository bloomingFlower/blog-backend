package models

import (
	"time"

	"gorm.io/gorm"
)

type AboutInfo struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string         `json:"name"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Contacts    []Contact      `json:"contacts" gorm:"foreignKey:AboutInfoID"`
	Sections    []Section      `json:"sections" gorm:"foreignKey:AboutInfoID"`
}

type Contact struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	AboutInfoID uint           `json:"-"`
	Icon        string         `json:"icon"`
	Label       string         `json:"label"`
	Link        string         `json:"link"`
}

type Section struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	AboutInfoID uint           `json:"-"`
	Title       string         `json:"title"`
	Icon        string         `json:"icon"`
	Items       []SectionItem  `json:"items" gorm:"foreignKey:SectionID"`
}

type SectionItem struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	SectionID   uint           `json:"-"`
	Title       string         `json:"title"`
	Institution string         `json:"institution,omitempty"`
	Company     string         `json:"company,omitempty"`
	Year        string         `json:"year,omitempty"`
	Period      string         `json:"period,omitempty"`
	Description string         `json:"description,omitempty"`
}
