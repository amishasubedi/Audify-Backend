package middleware

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Custom JWT claims struct
type CustomClaims struct {
	UserID string `json:"userId"`
	jwt.StandardClaims
}

// isAuthenticated middleware in Go
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

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		user, err := FindUserByIdAndToken(claims.UserID, tokenString)
		if err != nil || user == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access"})
			c.Abort()
			return
		}

		c.Set("user", user)
	} else {
		fmt.Println(err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	c.Next()
}

func FindUserByIdAndToken(userID, tokenString string) (*models.User, error) {
	var user models.User
	result := initializers.DB.Where("id = ? AND tokens = ?", userID, tokenString).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
