package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
* This method extracts and validates a file upload from an incoming request
 */
func FileUploadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Bad request",
			})
			return
		}
		defer file.Close()

		c.Set("filePath", header.Filename)
		c.Set("file", file)

		c.Next()
	}
}
