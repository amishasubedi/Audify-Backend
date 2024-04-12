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

	newUser.Password = models.HashPassword(newUser.Password)

	if result := initializers.DB.Create(&newUser); result.Error != nil {
		c.Error(result.Error)
		return
	}

	token := utils.GenerateToken(6)
	verificationRecord := models.UserEmailVerification{
		UserID:    newUser.ID,
		Token:     models.HashPassword(token),
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
* This method fetches and returns all registered users.
 */
func GetAllUsers(c *gin.Context) {
	var users []models.User
	if result := initializers.DB.Find(&users); result.Error != nil {
		c.Error(result.Error)
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
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
	token_hashed := models.HashPassword(token)

	verificationRecord := models.UserEmailVerification{
		UserID:    user.ID,
		Token:     token_hashed,
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
	result := initializers.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "User not found"})
		return
	}

	if !models.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid Password"})
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
		},
		"token": tokenString,
	})
}

/*
* Logout authenticated user
 */
func Signout(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	tokenStr := c.MustGet("token").(string)

	err := initializers.DB.Where("user_id = ? AND token = ?", userID, tokenStr).Delete(&models.Token{}).Error
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
* This method generates reset password link and send that link to user's email.
 */
func GeneratePasswordLink(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	result := initializers.DB.Where("email = ?", req.Email).First(&user)

	if result.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "User not found"})
		return
	}

	initializers.DB.Where("user_id = ?", user.ID).Delete(&models.UserPasswordReset{})

	token := utils.GenerateRandomHexString(36)

	passwordRecord := models.UserPasswordReset{
		UserID:    user.ID,
		Token:     token,
		CreatedAt: time.Now(),
	}

	if err := initializers.DB.Create(&passwordRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create password reset record"})
		return
	}

	link := os.Getenv("PASSWORD_RESET_LINK")
	resetLink := fmt.Sprintf("%s?token=%s&userId=%d", link, token, user.ID)

	emailOption := utils.Option{
		Email: user.Email,
		Link:  resetLink,
	}
	utils.SendForgetPasswordLink(emailOption)

	c.JSON(http.StatusOK, gin.H{"message": "Please check your email"})
}

/*
* This method verifies if the reset token entered by user is valid
 */
func IsValidResetToken(c *gin.Context) {
	// pass token and userid in the request body
	var req struct {
		UserID uint
		Token  string
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// find that reset token info using the user id from request body, handle error if not found
	var resetToken models.UserPasswordReset

	result := initializers.DB.Where("user_id = ?", req.UserID).First(&resetToken)
	if result.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Reset Token not found"})
		return
	}

	// check if the token passed and token stores matches
	matched, err := models.CompareToken(req.Token, resetToken.Token)

	if err != nil || !matched {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid Token"})
		return
	}

	// if it matches, say your token is valid
	c.JSON(http.StatusOK, gin.H{"message": "Your token is valid"})
}

/*
* This method updates users password
 */
func UpdatePassword(c *gin.Context) {
	var req struct {
		UserID   uint
		Password string
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User

	result := initializers.DB.Where("id = ?", req.UserID).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "User not found"})
		return
	}

	matched := models.CheckPasswordHash(req.Password, user.Password)
	if matched {
		c.JSON(http.StatusForbidden, gin.H{"error": "The new password should be different"})
		return
	}

	hashedPassword := models.HashPassword(req.Password)

	user.Password = hashedPassword

	if err := initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	var reset models.UserPasswordReset
	if err := initializers.DB.Where("user_id = ?", req.UserID).Delete(&reset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete verification token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

/*
* This method handles image uploading and saving in cloud, and also updating profile
 */
func UpdateProfile(c *gin.Context) {
	user, exists := c.Get("user")
	var imageURL, public_id string

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

	file, err := c.FormFile("picFile")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is missing"})
		return
	}

	filepath := file.Filename

	picFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open audio file"})
		return
	}
	defer picFile.Close()

	imageURL, public_id, err = utils.UploadToCloudinary(picFile, filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload audio file"})
		return
	}

	userModel.AvatarURL = imageURL
	userModel.AvatarPublicID = public_id

	if err := initializers.DB.Save(&userModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully", "user": userModel})

}
