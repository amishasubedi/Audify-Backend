package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type StringSlice []string

/*
* This method implements sql.Scanner interface and will be called by the database/sql package when scanning
* a column from the database
 */
func (ss *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*ss = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, ss)
}

/*
* This method implements the driver.Value interface which is called when saving value to database
 */
func (ss StringSlice) Value() (driver.Value, error) {
	if ss == nil {
		return nil, nil
	}

	return json.Marshal(ss)
}

type User struct {
	gorm.Model
	Name           string      `gorm:"column:name" validate:"required,min=3,max=20"`
	Email          string      `gorm:"column:email;unique" validate:"required,email"`
	Password       string      `gorm:"column:password" validate:"required,min=8"`
	AvatarURL      string      `gorm:"column:avatar_url" validate:"omitempty,url"`
	AvatarPublicID string      `gorm:"column:avatar_public_id" validate:"omitempty,alphanum"`
	Verified       bool        `gorm:"column:verified"`
	Favorites      []string    `gorm:"type:jsonb" validate:"omitempty,dive,required"`
	Followers      []string    `gorm:"type:jsonb" validate:"omitempty,dive,required"`
	Followings     []string    `gorm:"type:jsonb" validate:"omitempty,dive,required"`
	Tokens         StringSlice `gorm:"type:jsonb" validate:"omitempty,dive,required"`
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
