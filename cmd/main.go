package main

import (
	"backend/internal/initializers"
	"backend/internal/middleware"
	"backend/internal/routes"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectDatabase()
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(middleware.ErrorHandlingMiddleware())

	userRoutes := router.Group("/users")
	{
		routes.SetUserRoutes(userRoutes)
	}

	router.Run()
}
