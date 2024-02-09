package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"backend/internal/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

var Validate *validator.Validate

// Initialize the validator
func init() {
	Validate = validator.New()
}

/*
* This method creates a new user and initiates email verification
 */
func CreateUser(c *gin.Context) {
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

	token := utils.GenerateToken(6)
	//token_hash := models.HashPassword(token)

	emailVerification := models.UserEmailVerification{
		UserID:    newUser.ID,
		Token:     token,
		CreatedAt: time.Now(),
	}

	if err := initializers.DB.Create(&emailVerification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create email verification record"})
		return
	}

	// Send verification email
	profile := utils.Profile{
		Name:   newUser.Name,
		Email:  newUser.Email,
		UserID: fmt.Sprintf("%d", newUser.ID),
	}
	utils.SendVerificationMail(token, profile)

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

/*
* This method verifies a user's email using provided token
 */
func VerifyEmail(c *gin.Context) {
	var verificationToken models.UserEmailVerification

	if err := c.BindJSON(&verificationToken); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := initializers.DB.Where("user_id = ? AND token_hash = ?", verificationToken.UserID, models.HashPassword(verificationToken.Token)).First(&verificationToken)

	if result.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid Token"})
		return
	}

	// Update the User's verified status
	var user models.User
	updateResult := initializers.DB.Model(&user).Where("id = ?", verificationToken.UserID).Update("verified", true)
	if updateResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user verification status"})
		return
	}

	// Delete the verification token
	deleteResult := initializers.DB.Delete(&verificationToken)
	if deleteResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete verification token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your email is verified"})

}
