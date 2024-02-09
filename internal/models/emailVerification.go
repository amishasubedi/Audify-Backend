package models

import (
	"time"
)

type UserEmailVerification struct {
	VerificationID uint      `gorm:"primaryKey"`
	UserID         uint      `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TokenHash      string    `gorm:"not null"`
	CreatedAt      time.Time `gorm:"default:current_timestamp"`
	User           User      `gorm:"foreignKey:UserID"`
}
