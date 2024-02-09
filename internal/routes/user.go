package routes

import (
	"backend/internal/controllers"

	"github.com/gin-gonic/gin"
)

func SetUserRoutes(router *gin.RouterGroup) {
	router.POST("/", controllers.CreateUser)
	router.GET("/", controllers.GetAllUsers)

}
