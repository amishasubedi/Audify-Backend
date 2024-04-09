package routes

import (
	"backend/internal/controllers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetPlaylistRoutes(router *gin.RouterGroup) {
	router.POST("/create", middleware.IsAuthenticated, controllers.CreatePlaylist)
	router.POST("/add", middleware.IsAuthenticated, controllers.AddToPlaylist)

	router.POST("/update-playlist", middleware.IsAuthenticated, controllers.UpdatePlaylist)
	router.POST("/remove/", middleware.IsAuthenticated, controllers.RemoveFromPlaylist)
	router.DELETE("/delete/:playlistId", middleware.IsAuthenticated, controllers.DeletePlaylist)

	router.GET("/:playlistId", middleware.IsAuthenticated, controllers.GetAudiosByPlaylist)
	router.GET("/detail/:playlistId", middleware.IsAuthenticated, controllers.GetPlaylistDetailsByID)

}
