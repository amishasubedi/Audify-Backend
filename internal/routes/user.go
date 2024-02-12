package routes

import (
	"backend/internal/controllers"

	"github.com/gin-gonic/gin"
)

func SetUserRoutes(router *gin.RouterGroup) {
	router.POST("/", controllers.CreateUser)
	router.POST("/verify", controllers.VerifyEmail)
	router.POST("/re-verify", controllers.ReVerifyEmail)
	router.POST("/sign-in", controllers.Signin)
	router.GET("/", controllers.GetAllUsers)
}
