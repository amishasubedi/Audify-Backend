package routes

import (
	"backend/internal/controllers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetFavoriteRoutes(router *gin.RouterGroup) {
	router.POST("/add", middleware.IsAuthenticated, controllers.AddToFavorite)
	router.POST("/delete", middleware.IsAuthenticated, controllers.DeleteFromFavorite)
	router.GET("/my-favorite", middleware.IsAuthenticated, controllers.GetAllFavorites)
}
