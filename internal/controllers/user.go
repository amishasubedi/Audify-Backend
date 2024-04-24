package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"backend/internal/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
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
		c.Error(err)
		return
	}

	if validationErr := Validate.Struct(newUser); validationErr != nil {
		c.Error(validationErr)
		return
	}

	if result := initializers.DB.Create(&newUser); result.Error != nil {
		c.Error(result.Error)
		return
	}

	token := utils.GenerateToken(6)
	verificationRecord := models.UserEmailVerification{
		UserID:    newUser.ID,
		Token:     token,
		CreatedAt: time.Now(),
	}

	if result := initializers.DB.Create(&verificationRecord); result.Error != nil {
		c.Error(result.Error)
		return
	}

	profile := utils.Profile{Name: newUser.Name, Email: newUser.Email, UserID: fmt.Sprintf("%d", newUser.ID)}
	if err := utils.SendVerificationMail(token, profile); err != nil {
		log.Printf("Error sending verification mail: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"data": newUser})
}

/*
* This method verifies a user's email using provided token.
 */
func VerifyEmail(c *gin.Context) {
	var req models.VerifyEmail

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	var verificationToken models.UserEmailVerification
	result := initializers.DB.Where("user_id = ?", req.UserID).First(&verificationToken)
	if result.Error == gorm.ErrRecordNotFound {
		c.Error(result.Error)
	}

	if matched, err := models.CompareToken(verificationToken.Token, req.Token); err != nil || !matched {
		if err == nil {
			err = fmt.Errorf("token mismatch")
		}
		c.Error(err)
		return
	}

	var user models.User
	updateResult := initializers.DB.Model(&user).Where("id = ?", verificationToken.UserID).Update("verified", true)
	if updateResult.Error != nil {
		c.Error(updateResult.Error)
		return
	}

	deleteResult := initializers.DB.Delete(&verificationToken)
	if deleteResult.Error != nil {
		c.Error(deleteResult.Error)
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
		c.Error(err)
		return
	}

	var user models.User
	result := initializers.DB.Where("id = ?", req.UserID).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		c.Error(result.Error)
		return
	}

	token := utils.GenerateToken(6)

	verificationRecord := models.UserEmailVerification{
		UserID:    user.ID,
		Token:     token,
		CreatedAt: time.Now(),
	}

	if result := initializers.DB.Create(&verificationRecord); result.Error != nil {
		c.Error(result.Error)
		return
	}

	profile := utils.Profile{
		Name:   user.Name,
		Email:  user.Email,
		UserID: fmt.Sprintf("%d", user.ID),
	}

	if err := utils.SendVerificationMail(token, profile); err != nil {
		log.Printf("Error sending verification mail: %v", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Please check your email"})
}

/**
 * Sign in user
 */
func Signin(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := initializers.DB.Select("ID", "Password", "Email", "Name", "Verified", "AvatarURL").
		Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "User not found"})
		return
	}

	if !models.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign the token"})
		return
	}

	newToken := models.Token{
		UserID:    user.ID,
		Token:     tokenString,
		Type:      "auth",
		ExpiresAt: time.Now().Add(time.Hour * 72),
	}

	if err := initializers.DB.Create(&newToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save new token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"verified": user.Verified,
			"avatar":   user.AvatarURL,
			"admin":    user.IsAdmin,
		},
		"token": tokenString,
	})
}

/*
* Logout authenticated user
 */
func Signout(c *gin.Context) {
	user, exists := c.Get("user")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	userModel, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User casting error"})
		return
	}

	tokenStr := c.PostForm("token")
	if tokenStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token not provided"})
		return
	}

	err := initializers.DB.Where("user_id = ? AND token = ?", userModel.ID, tokenStr).Delete(&models.Token{}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign out"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Signed out successfully"})
}

func SendProfile(c *gin.Context) {
	user, exists := c.Get("user")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": user})
}

/*
* This method updates users password
 */
func UpdatePassword(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	result := initializers.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if models.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The new password must be different from the current one"})
		return
	}

	user.Password = req.Password
	if err := initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

/*
* This method handles image uploading and saving in cloud, and also updating profile
 */
func UpdateProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	userModel, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	name := c.PostForm("name")
	bio := c.PostForm("bio")
	if name != "" || bio != "" {
		if len(name) < 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid name, must be at least 3 characters long"})
			return
		}
		userModel.Name = name
		userModel.Bio = bio
	}

	file, fileErr := c.FormFile("picFile")
	if fileErr == nil {
		if userModel.AvatarPublicID != "" {
			err := utils.DestroyImage(userModel.AvatarPublicID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to destroy existing image"})
				return
			}
		}

		filepath := file.Filename
		picFile, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		defer picFile.Close()

		imageURL, publicID, err := utils.UploadToCloudinary(picFile, filepath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
			return
		}

		userModel.AvatarURL = imageURL
		userModel.AvatarPublicID = publicID
	} else if fileErr != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error retrieving file"})
		return
	}

	if err := initializers.DB.Save(&userModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// GetRecommendedUsers returns a list of users who have uploaded at least 3 songs
func GetRecommendedUsers(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	userModel, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	var users []models.User
	result := initializers.DB.
		Joins("JOIN audios ON audios.owner = users.id").
		Where("users.id != ?", userModel.ID).
		Group("users.id").
		Having("count(audios.id) >= ?", 3).
		Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	userProfiles := make([]map[string]interface{}, len(users))
	for i, user := range users {
		userProfiles[i] = map[string]interface{}{
			"id":     user.ID,
			"name":   user.Name,
			"avatar": user.AvatarURL,
		}
	}

	c.JSON(http.StatusOK, gin.H{"recommendedUsers": userProfiles})
}
