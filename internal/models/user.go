package models

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name           string   `gorm:"column:name" validate:"required,min=3,max=20"`
	Email          string   `gorm:"column:email;unique" validate:"required,email"`
	Password       string   `gorm:"column:password_hash" validate:"required,min=8"`
	AvatarURL      string   `gorm:"column:avatar_url" validate:"omitempty,url"`
	AvatarPublicID string   `gorm:"column:avatar_public_id" validate:"omitempty,alphanum"`
	Verified       bool     `gorm:"column:verified"`
	Favorites      []string `gorm:"type:jsonb" validate:"omitempty,dive,required"`
	Followers      []string `gorm:"type:jsonb" validate:"omitempty,dive,required"`
	Followings     []string `gorm:"type:jsonb" validate:"omitempty,dive,required"`
	Tokens         []string `gorm:"type:jsonb" validate:"omitempty,dive,required"`
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
