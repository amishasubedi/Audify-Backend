package models

import (
	"time"
)

type UserEmailVerification struct {
	VerificationID uint      `gorm:"primaryKey"`
	Token          string    `gorm:"column:token_hash;not null"`
	CreatedAt      time.Time `gorm:"default:current_timestamp"`
	UserID         uint      `gorm:"foreignKey:UserID"`
}
