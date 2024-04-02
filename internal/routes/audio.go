package routes

import (
	"backend/internal/controllers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetAudioRoutes(router *gin.RouterGroup) {
	router.POST("/create", middleware.IsAuthenticated, middleware.FileParserMiddleware(), controllers.CreateAudio)
	router.PATCH("/:audioId", middleware.IsAuthenticated, middleware.FileParserMiddleware(), controllers.UpdateAudio)
	router.GET("/recommendation", middleware.IsAuthenticated, controllers.GetSuggestionsList)
	router.GET("/category", controllers.FilterByMood)

	router.GET("/", controllers.GetLatestAudios)
	router.GET("/latest-uploads", controllers.GetLatestUploads)
}
