package models

type Post struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Image   string `json:"image"`
	Tags    string `json:"tags"`
	UserID  uint   `json:"user_id"`
	User    User   `json:"user" gorm:"foreignKey:UserID"`
}
