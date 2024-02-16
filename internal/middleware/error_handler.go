package middleware

import (
	"net/http"

	response "backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		for _, err := range c.Errors {
			var code int
			var message string
			var data interface{}

			switch e := err.Err.(type) {
			case validator.ValidationErrors:
				code = http.StatusBadRequest
				message = "Validation failed"
				data = e
			case *validator.InvalidValidationError:
				code = http.StatusBadRequest
				message = "Invalid data"
			case *gin.Error:
				code = http.StatusInternalServerError
				message = "Internal Server Error"
				if e.Type == gin.ErrorTypePublic {
					message = e.Error()
				}
			default:
				code = http.StatusInternalServerError
				message = "Internal Server Error"
			}

			response := response.NewErrorResponse(code, message, data)
			c.JSON(code, response)
			return
		}
	}
}
