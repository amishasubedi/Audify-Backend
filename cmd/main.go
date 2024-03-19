package main

import (
	"backend/internal/initializers"
	"backend/internal/middleware"
	"backend/internal/routes"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectDatabase()
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(middleware.EnableCors())
	router.Use(middleware.ErrorHandlingMiddleware())

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
