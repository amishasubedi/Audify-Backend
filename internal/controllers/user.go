package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
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

	tokensJSON, err := json.Marshal(append(user.Tokens, tokenString)) // multiple sign in token not appended
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process tokens"})
		return
	}

	if err := initializers.DB.Model(&user).Update("tokens", tokensJSON).Error; err != nil {
		fmt.Printf("Error updating user tokens in database: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update user tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"verified":   user.Verified,
			"avatar":     user.AvatarURL,
			"tokens":     len(user.Tokens),
			"followers":  len(user.Followers),
			"followings": len(user.Followings),
		},
		"token": tokenString,
	})
}

/*
* Logout authenticated user
 */
func Signout(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	token := c.MustGet("token").(string)
	fromAll := c.Query("fromAll")

	if fromAll == "yes" {
		initializers.DB.Model(&models.User{}).Where("id = ?", userID).Update("tokens", []string{})
	} else {
		var user models.User
		if err := initializers.DB.First(&user, userID).Error; err == nil {
			newTokens := []string{}
			for _, t := range user.Tokens {
				if t != token {
					newTokens = append(newTokens, t)
				}
			}
			initializers.DB.Model(&user).Update("tokens", newTokens)
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
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
		c.Error(err)
		return
	}

	var user models.User
	result := initializers.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		c.Error(result.Error)
		return
	}

	//var passwordSchema models.UserPasswordReset

	// delete that users info from passwordResetSchema only if it exists
	// output := initializers.DB.Where("email = ?", user.Email).First(&passwordSchema)
	// if output.Error != nil {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "User not found 1"})
	// }

	// deleteResult := initializers.DB.Delete(&passwordSchema)
	// if deleteResult.Error != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete verification token"})
	// 	return
	// }

	// generate token
	token := utils.GenerateRandomHexString(36)

	// create new passwordReset with that token
	passwordRecord := models.UserPasswordReset{
		UserID:    user.ID,
		Token:     token,
		CreatedAt: time.Now(),
	}

	if result := initializers.DB.Create(&passwordRecord); result.Error != nil {
		c.Error(result.Error)
		return
	}

	// generate resetLink
	link := os.Getenv("PASSWORD_RESET_LINK")
	resetLink := fmt.Sprintf("%s?token=%s&userId=%d", link, token, user.ID)

	option := utils.Option{
		Email: user.Email,
		Link:  resetLink,
	}

	// send that link to users email
	utils.SendForgetPasswordLink(option)

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

	fmt.Print("Request body token", req.Token)
	fmt.Print("Database TOken", resetToken.Token)

	// check if the token passed and token stores matches
	matched, err := models.CompareToken(req.Token, resetToken.Token)

	if err != nil || !matched {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid Token"})
		return
	}

	// if it matches, say your token is valid
	c.JSON(http.StatusOK, gin.H{"message": "Your token is valid"})
}

func UpdatePassword(c *gin.Context) {
	// pass user id and password in request body
	var req struct {
		UserID   uint
		Password string
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// find that user in the database, handle error
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

func UpdateProfile(c *gin.Context) {
	userID := c.Param("userID")

	cld := utils.CloudSetup()

	name := c.PostForm("name")
	if name == "" || len(name) < 3 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Name must be at least 3 characters long"})
		return
	}

	var user models.User
	if err := initializers.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error, user not found"})
		return
	}

	user.Name = name

	file, err := c.FormFile("avatar")
	if err == nil && file != nil {
		tempFilePath := fmt.Sprintf("/tmp/%s", file.Filename)
		if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
			return
		}

		url, err := utils.UploadFileToCloudinary(cld, tempFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload avatar to Cloudinary"})
			return
		}

		user.AvatarURL = url
	}

	if err := initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully", "user": user})
}
