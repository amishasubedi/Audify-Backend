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
	router.PATCH("/update-password", controllers.UpdatePassword)

	// profile management route
	router.PATCH("/check", middleware.IsAuthenticated, middleware.FileParserMiddleware(), controllers.UpdateProfile)

	router.GET("/recommendation", middleware.IsAuthenticated, controllers.GetRecommendedUsers)

	// admin
	router.GET("/all-users", controllers.GetAllUsers)
	router.GET("/contents/playlists/:userId", middleware.IsAuthenticated, middleware.IsAdmin, controllers.GetPlaylistsByUser)
	router.DELETE("/delete/playlist/:playlistId", middleware.IsAuthenticated, middleware.IsAdmin, controllers.DeletePlaylistById)
	router.DELETE("/delete/audio/:audioId", middleware.IsAuthenticated, middleware.IsAdmin, controllers.DeleteAudioById)
}
