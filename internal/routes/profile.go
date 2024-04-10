package routes

import (
	"backend/internal/controllers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetProfileRoutes(router *gin.RouterGroup) {
	router.GET("/user/:userId", controllers.GetPublicProfile)

	router.GET("/my-songs", middleware.IsAuthenticated, controllers.GetPersonalUploads)
	router.GET("/my-playlist", middleware.IsAuthenticated, controllers.GetPersonalPlaylist)

	router.POST("/follow/:followingId", middleware.IsAuthenticated, controllers.FollowUser)

}
