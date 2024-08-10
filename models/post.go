package models

import "time"

type Post struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	File      string     `json:"file"`
	Tags      string     `json:"tags"`
	Category  string     `json:"category"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"autoUpdateTime:false"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at"`
	Hidden    bool       `json:"hidden"`
	UserID    uint       `json:"user_id"`
	User      User       `json:"user" gorm:"foreignKey:UserID"`
}
