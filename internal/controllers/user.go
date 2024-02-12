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
	"gorm.io/gorm"
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

	// generate token
	token := utils.GenerateToken(6)
	token_hashed := models.HashPassword(token)

	// create email verification record on database
	verificationRecord := models.UserEmailVerification{
		UserID:    newUser.ID,
		Token:     token_hashed,
		CreatedAt: time.Now(),
	}

	if result := initializers.DB.Create(&verificationRecord); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	profile := utils.Profile{
		Name:   newUser.Name,
		Email:  newUser.Email,
		UserID: fmt.Sprintf("%d", newUser.ID),
	}

	// send verification email
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
	var req models.VerifyEmail

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var verificationToken models.UserEmailVerification

	// find userid in email verification associated with req bpdy
	result := initializers.DB.Where("user_id = ?", req.UserID).First(&verificationToken)
	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(http.StatusForbidden, gin.H{"error": "Token not found"})
		return
	}

	// check if the req body token and token stored in database are same
	matched, err := models.CompareToken(verificationToken.Token, req.Token)
	if err != nil || !matched {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid Token"})
		return
	}

	// if it matches, find the user by id and update verify to true in users table
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

/**
 * This method resends a verification token to the user's email.
 */
func ReVerifyEmail(c *gin.Context) {
	var req models.ReVerifyEmail

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User

	// find userid in email verification associated with req bpdy
	result := initializers.DB.Where("id = ?", req.UserID).First(&user)
	//fmt.Println("User Info: ", verificationID)
	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid User"})
		return
	}

	// generate new token
	token := utils.GenerateToken(6)
	token_hashed := models.HashPassword(token)

	// create email verification record on database
	verificationRecord := models.UserEmailVerification{
		UserID:    user.ID,
		Token:     token_hashed,
		CreatedAt: time.Now(),
	}

	if result := initializers.DB.Create(&verificationRecord); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	profile := utils.Profile{
		Name:   user.Name,
		Email:  user.Email,
		UserID: fmt.Sprintf("%d", user.ID),
	}

	// send verification email
	utils.SendVerificationMail(token, profile)

	c.JSON(http.StatusOK, gin.H{"message": "Please check your email"})
}
