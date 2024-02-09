package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

var Validate *validator.Validate

// Initialize the validator
func init() {
	Validate = validator.New()
}

func CreateUser(c *gin.Context) {
	// Retrieve the validated UserRequestBody from the context
	var newUser models.User

	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validationErr := Validate.Struct(newUser)

	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	password := models.HashPassword(newUser.Password)
	newUser.Password = password

	result := initializers.DB.Create(&newUser)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// token := utils.GenerateToken(6)

	// emailVerification := models.UserEmailVerification{
	// 	UserID:    newUser.ID,
	// 	TokenHash: token,
	// 	CreatedAt: time.Now(),
	// }

	// if err := initializers.DB.Create(&emailVerification).Error; err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create email verification record"})
	// 	return
	// }

	// // Send verification email
	// profile := utils.Profile{
	// 	Name:   newUser.Name,
	// 	Email:  newUser.Email,
	// 	UserID: fmt.Sprintf("%d", newUser.ID),
	// }
	// utils.SendVerificationMail(token, profile)

	c.JSON(http.StatusOK, gin.H{"data": newUser})
}

func GetAllUsers(c *gin.Context) {
	var users []models.User
	result := initializers.DB.Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}
