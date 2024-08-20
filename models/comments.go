package models

import (
	"time"

	"gorm.io/gorm"
)

// Comment represents a comment on a post
type Comment struct {
	gorm.Model
	PostID    uint      `json:"post_id"`
	UserID    *uint     `json:"user_id"`
	ParentID  *uint     `json:"parent_id"`
	Content   string    `json:"content"`
	Votes     []Vote    `json:"votes" gorm:"foreignKey:CommentID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Parent    *Comment  `json:"parent" gorm:"foreignKey:ParentID"`
	Children  []Comment `json:"children" gorm:"foreignKey:ParentID"`
	IPAddress string    `json:"ip_address"`
}

// Vote represents a vote (emoji reaction) on a comment
type Vote struct {
	gorm.Model
	CommentID uint      `json:"comment_id"`
	UserID    *uint     `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	IPAddress string    `json:"ip_address"`
}
