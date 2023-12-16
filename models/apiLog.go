package models

import (
	"time"
)

type APILog struct {
	ID                 uint          `gorm:"primaryKey" json:"id"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
	DeletedAt          *time.Time    `gorm:"index" json:"deleted_at"`
	RequestMethod      string        `json:"request_method"`
	RequestURL         string        `json:"request_url"`
	RequestIP          string        `json:"request_ip"`
	ResponseStatusCode int           `json:"response_status_code"`
	ResponseTime       time.Duration `json:"response_time"`
	UserID             uint          `json:"user_id"`
}
