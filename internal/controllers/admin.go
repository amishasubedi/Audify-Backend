package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func DeletePlaylistById(c *gin.Context) {
	playlistID := c.Param("playlistId")

	var playlist models.Playlist

	if err := initializers.DB.First(&playlist, playlistID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audio not found"})
		return
	}

	if err := initializers.DB.Delete(&playlist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist successfully deleted"})
}

func GetPlaylistsByUser(c *gin.Context) {
	userID := c.Param("userId")

	var playlists []models.Playlist
	if err := initializers.DB.Where("owner_id = ?", userID).Find(&playlists).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlists not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
}
