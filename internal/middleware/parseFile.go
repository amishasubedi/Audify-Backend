package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
* This method parse the incoming file from request body
 */
func FileParser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if contentType := c.GetHeader("Content-Type"); contentType != "" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Only accepts form data"})
			c.Abort()
			return
		}

		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing form data"})
			c.Abort()
			return
		}

		form := c.Request.MultipartForm
		fields := map[string]string{}
		if form != nil {
			for key, value := range form.Value {
				fields[key] = value[0]
			}
			for key, files := range form.File {
				file := files[0]
				fields[key] = file.Filename
			}
		}
		c.Set("parsedFields", fields)

		c.Next()
	}
}
