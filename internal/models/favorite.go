package models

type Favorite struct {
	UserID  uint  `gorm:"primaryKey"`
	AudioID uint  `gorm:"primaryKey"`
	User    User  `gorm:"foreignKey:UserID"`
	Audio   Audio `gorm:"foreignKey:AudioID"`
}
