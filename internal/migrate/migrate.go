package main

import (
	"backend/internal/initializers"
	"backend/internal/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectDatabase()
}

func main() {
	initializers.DB.AutoMigrate(&models.User{})
	initializers.DB.AutoMigrate(&models.UserEmailVerification{})
	initializers.DB.AutoMigrate(&models.UserPasswordReset{})
	initializers.DB.AutoMigrate(&models.Audio{})
	initializers.DB.AutoMigrate(&models.Playlist{})
}
