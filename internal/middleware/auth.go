package middleware

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

/*
* Custom JWT Claim Struct
 */
type CustomClaims struct {
	UserID uint `json:"userId"`
	jwt.StandardClaims
}

func FindUserByIdAndToken(userID uint, tokenString string) (*models.User, error) {
	var token models.Token
	if err := initializers.DB.Where("user_id = ? AND token = ?", userID, tokenString).First(&token).Error; err != nil {
		return nil, err
	}

	var user models.User
	if err := initializers.DB.Preload("Tokens").First(&user, userID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

/*
* This method extracts the token from authorization header, and validate the token
 */
func IsAuthenticated(c *gin.Context) {
	authorization := c.GetHeader("Authorization")
	if authorization == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access!"})
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authorization, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		user, err := FindUserByIdAndToken(claims.UserID, tokenString)
		if err != nil || user == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access here"})
			c.Abort()
			return
		}

		c.Set("user", user)
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	c.Next()
}

func IsAdmin(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access!"})
		c.Abort()
		return
	}

	userModel, ok := user.(*models.User)
	if !ok || !userModel.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required!"})
		c.Abort()
		return
	}

	c.Next()
}
