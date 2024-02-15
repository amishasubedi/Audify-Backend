package routes

import (
	"backend/internal/controllers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetUserRoutes(router *gin.RouterGroup) {

	// Authentication Routes
	router.POST("/sign-in", controllers.Signin)
	router.POST("/sign-up", controllers.CreateUser)
	router.POST("/logout", middleware.IsAuthenticated, controllers.Signout)
	router.POST("/is-auth", middleware.IsAuthenticated, controllers.SendProfile)

	// Email Verification Routes
	router.POST("/verify", controllers.VerifyEmail)
	router.POST("/re-verify", controllers.ReVerifyEmail)

	// Password management routes
	router.POST("/password-reset", controllers.GeneratePasswordLink)
	router.POST("/is-valid-token", controllers.IsValidResetToken)
	router.POST("/update-password", controllers.UpdatePassword)

	// profile management route
	router.POST("/update-profile", controllers.UpdateProfile)
	router.GET("/", controllers.GetAllUsers)
}
