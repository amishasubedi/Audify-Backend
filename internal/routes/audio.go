package routes

import (
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAudioRoutes(router *gin.RouterGroup) {
	router.POST("/create", middleware.IsAuthenticated, middleware.FileUploadMiddleware(), CreateAudio)
}
