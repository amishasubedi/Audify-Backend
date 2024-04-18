package models

import (
	"backend/internal/initializers"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name           string   `gorm:"column:name;validate:'required,min=3,max=20'"`
	Email          string   `gorm:"column:email;unique;validate:'required,email'"`
	Password       string   `gorm:"column:password;validate:'required,min=8'"`
	Bio            string   `gorm:"column:bio;"`
	AvatarURL      string   `gorm:"column:avatar_url;validate:'omitempty,url'"`
	AvatarPublicID string   `gorm:"column:avatar_public_id;validate:'omitempty,alphanum'"`
	Verified       bool     `gorm:"column:verified"`
	IsAdmin        bool     `gorm:"column:is_admin"`
	Favorites      []*Audio `gorm:"many2many:user_favorites;"`
	Tokens         []*Token `gorm:"foreignKey:UserID"`
}

type User_Relations struct {
	gorm.Model
	FollowerID  uint `gorm:"not null"`
	FollowingID uint `gorm:"not null"`
	Follower    User `gorm:"foreignKey:FollowerID"`
	Following   User `gorm:"foreignKey:FollowingID"`
}

type Token struct {
	gorm.Model
	UserID    uint      `gorm:"index"`
	Token     string    `gorm:"column:token;unique"`
	Type      string    `gorm:"column:type"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
}

// save user details
func (user *User) Save() (*User, error) {
	err := initializers.DB.Create(&user).Error
	if err != nil {
		return &User{}, err
	}
	return user, nil
}

// generate encrypted password
func (user *User) BeforeSave(*gorm.DB) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(passwordHash)
	return nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
