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
	router.POST("/password-reset", controllers.GeneratePasswordLink)
	router.POST("/is-valid-token", controllers.IsValidResetToken)

	router.GET("/", controllers.GetAllUsers)
}
