package main

import (
	"backend/internal/initializers"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	//initializers.ConnectDatabase()
}

func RunMigrations() {
	err := initializers.DB.AutoMigrate(
		&models.User{},
		&models.UserEmailVerification{},
		&models.UserPasswordReset{},
		&models.Audio{},
		&models.Playlist{},
		&models.Token{},
		&models.Favorite{},
		&models.User_Relations{},
	)

	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(middleware.EnableCors())
	router.Use(middleware.ErrorHandlingMiddleware())

	if os.Getenv("RUN_MIGRATIONS") == "true" {
		RunMigrations()
	}

	userRoutes := router.Group("/users")
	{
		routes.SetUserRoutes(userRoutes)
	}

	audioRoutes := router.Group("/audio")
	{
		routes.SetAudioRoutes(audioRoutes)
	}

	playlistRoutes := router.Group("/playlist")
	{
		routes.SetPlaylistRoutes(playlistRoutes)
	}

	favoriteRoutes := router.Group("/favorite")
	{
		routes.SetFavoriteRoutes(favoriteRoutes)
	}
	historyRoutes := router.Group("/history")
	{
		routes.SetHistoryRoutes(historyRoutes)
	}
	profileRoutes := router.Group("/profile")
	{
		routes.SetProfileRoutes(profileRoutes)
	}

	router.Run()
}
