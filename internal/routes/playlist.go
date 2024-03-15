package routes

import (
	"backend/internal/controllers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetPlaylistRoutes(router *gin.RouterGroup) {
	router.POST("/create", middleware.IsAuthenticated, controllers.CreatePlaylist)
	router.POST("/update-playlist", middleware.IsAuthenticated, controllers.UpdatePlaylist)

	router.GET("/:playlistId", middleware.IsAuthenticated, controllers.GetAudiosByPlaylist)
	router.GET("/my-playlist", middleware.IsAuthenticated, controllers.GetPlaylistByProfile)

	router.DELETE("/", middleware.IsAuthenticated, controllers.DeletePlaylist)
}
