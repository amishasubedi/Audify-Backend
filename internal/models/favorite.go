package models

import "gorm.io/gorm"

type Favorite struct {
	gorm.Model
	Owner uint             `gorm:"column:owner_id;foreignKey:UserID"`
	Songs JSONIntegerArray `gorm:"column:songs;type:json;default:[]"`
}
