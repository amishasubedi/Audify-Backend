package routes

import (
	"backend/internal/controllers"

	"github.com/gin-gonic/gin"
)

func SetProfileRoutes(router *gin.RouterGroup) {
	router.GET("/user/:userId", controllers.GetPublicProfile)
	router.GET("/playlist/:userId", controllers.GetPublicPlaylists)

}
