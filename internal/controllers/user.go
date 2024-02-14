package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"backend/internal/utils"
	"fmt"
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
* This method verifies a user's email using provided token.
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access"})
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

	// Update user's tokens
	user.Tokens = append(user.Tokens, tokenString) // not appending ?

	output := initializers.DB.Save(&user)
	if output.Error != nil {
		fmt.Println("Error saving user to database:", result.Error)
	}
	c.JSON(http.StatusOK, gin.H{
		"profile": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"verified":   user.Verified,
			"avatar":     user.AvatarURL,
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

	// find user associated with that email
	var user models.User
	result := initializers.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "User with that email not found"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
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

	// find that user in user, handle error

	// compare password - if same - error - "new pwd should be different"

	// set users password to password passed in requets body

	// save that in db

	// delete info of tokens in passwordReset Schema

	// say passowrd change successfully
}

func UpdateProfile(c *gin.Context) {
	// file storage ?
}
