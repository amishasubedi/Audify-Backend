package models

import (
	"gorm.io/gorm"
)

type Audio struct {
	gorm.Model
	Title         string   `gorm:"column:name" validate:"required,min=10,max=200"`
	About         string   `gorm:"column:about" validate:"required max=1000"`
	Owner         uint     `gorm:"foreignKey:UserID"`
	AudioURL      string   `gorm:"column:audio_url" validate:"omitempty,url"`
	AudioPublicID string   `gorm:"column:audio_public_id" validate:"omitempty,alphanum"`
	CoverURL      string   `gorm:"column:cover_url" validate:"omitempty,url"`
	CoverPublicID string   `gorm:"column:cover_public_id" validate:"omitempty,alphanum"`
	Likes         []string `gorm:"type:jsonb" validate:"omitempty,dive,required"`
	Category      string   `gorm:"column:category" validate:"required"`
}
