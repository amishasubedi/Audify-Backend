package middleware

import (
	"net/http"

	response "backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

/*
* ErrorHandlingMiddleware intercepts errors added to the Gin context and processes them to return standardized error responses.
 */
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				processError(c, e.Err)
				return
			}
		}
	}
}

/*
* processError categorizes and formats errors before sending them as HTTP responses.
 */
func processError(c *gin.Context, err error) {
	var code int
	var message string
	var data interface{}

	switch e := err.(type) {
	case validator.ValidationErrors:
		code = http.StatusBadRequest
		message = "Validation failed"
		detailedErrors := make(map[string]string)
		for _, ve := range e {
			field := ve.Field()
			tag := ve.Tag()
			detailedErrors[field] = getValidationErrorMessage(field, tag, ve.Param())
		}
		data = detailedErrors
	case *validator.InvalidValidationError:
		code = http.StatusBadRequest
		message = "Invalid data structure"
	default:
		code = http.StatusInternalServerError
		message = "Internal Server Error"
	}

	resp := response.NewErrorResponse(code, message, data)
	c.JSON(code, resp)
}

/*
* getValidationErrorMessage generates readable error messages based on validation tags.
 */
func getValidationErrorMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return "Invalid email format"
	case "min":
		return field + " must be at least " + param + " characters long"
	default:
		return field + " is not valid"
	}
}
