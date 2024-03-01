package middleware

import (
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
* This method extracts and validates a file upload from an incoming request
 */
func FileParserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if contentType := c.GetHeader("Content-Type"); !startsWith(contentType, "multipart/form-data") {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Only accepts form data"})
			c.Abort()
			return
		}

		if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32 MB size
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request, error parsing form data"})
			c.Abort()
			return
		}

		fields := make(map[string]string)
		for key, value := range c.Request.Form {
			if len(value) > 0 {
				fields[key] = value[0]
			}
		}

		c.Set("fields", fields)

		files := make(map[string]*multipart.FileHeader)
		for key, fileHeaders := range c.Request.MultipartForm.File {
			if len(fileHeaders) > 0 {
				files[key] = fileHeaders[0]
			}
		}

		c.Set("files", files)

		c.Next()
	}
}

func startsWith(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}
