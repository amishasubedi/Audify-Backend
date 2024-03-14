package models

import (
	"gorm.io/gorm"
)

type Playlist struct {
	gorm.Model
	Title      string   `gorm:"column:title;validate:min=10,max=200"`
	Owner      uint     `gorm:"column:owner_id;foreignKey:UserID"`
	Songs      []string `gorm:"type:jsonb;validate:omitempty,dive,required"`
	Visibility string   `gorm:"column:visibility;default:public;validate:oneof=public private auto"`
}
