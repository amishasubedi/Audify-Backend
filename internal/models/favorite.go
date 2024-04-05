package models

import "gorm.io/gorm"

type Favorite struct {
	gorm.Model
	Owner  uint    `gorm:"column:owner_id;foreignKey:UserID"`
	Audios []Audio `gorm:"many2many:favorite_audios;"`
}
