package models

import (
	"time"
)

type UserPasswordReset struct {
	ResetID   uint      `gorm:"primaryKey"`
	Token     string    `gorm:"column:token;not null"`
	CreatedAt time.Time `gorm:"default:current_timestamp"`
	UserID    uint      `gorm:"foreignKey:UserID"`
}
