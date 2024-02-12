package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserEmailVerification struct {
	VerificationID uint      `gorm:"primaryKey"`
	Token          string    `gorm:"column:token;not null"`
	CreatedAt      time.Time `gorm:"default:current_timestamp"`
	UserID         uint      `gorm:"foreignKey:UserID"`
}

type VerifyEmail struct {
	Token  string `gorm:"column:token_hash;not null"`
	UserID uint   `gorm:"foreignKey:UserID"`
}

type ReVerifyEmail struct {
	UserID uint `gorm:"foreignKey:UserID"`
}

// CompareToken compares a plaintext token against a hashed token.
func CompareToken(hashedToken, token string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedToken), []byte(token))
	return err == nil, err
}
