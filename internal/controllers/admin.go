package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DeleteUser(c *gin.Context) {
	userID := c.Param("userId")

	var user models.User
	if err := initializers.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := initializers.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User successfully deleted"})
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

func DeleteAudioById(c *gin.Context) {
	audioID := c.Param("audioId")

	var audio models.Audio
	if err := initializers.DB.First(&audio, audioID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audio not found"})
		return
	}

	if err := initializers.DB.Delete(&audio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting audio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Audio successfully deleted"})
}
