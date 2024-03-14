package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
* This method creates a new playlist
 */
func CreatePlaylist(c *gin.Context) {
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

	title := c.PostForm("title")
	resId := c.PostForm("resId")
	visibility := c.PostForm("visibility")

	if title == "" || visibility == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	newPlaylist := models.Playlist{
		Title:      title,
		Owner:      userModel.ID,
		Visibility: visibility,
	}

	if resId != "" {
		var audioModel models.Audio
		if err := initializers.DB.Where("id = ?", resId).First(&audioModel).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Audio not found"})
			return
		}
		newPlaylist.Songs = append(newPlaylist.Songs, resId)
	}

	if err := initializers.DB.Create(&newPlaylist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Playlist created successfully",
		"playlist": gin.H{
			"id":         newPlaylist.ID,
			"title":      newPlaylist.Title,
			"visibility": newPlaylist.Visibility,
			"songs":      newPlaylist.Songs,
		},
	})
}

/*
* This method fetches the audio details of playlist by id
 */
func GetAudiosByPlaylist(c *gin.Context) {
	playlistId := c.Param("playlistId")

	var playlist models.Playlist

	if err := initializers.DB.Preload("Songs").Where("id = ?", playlistId).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	var audios []models.Audio

	for _, songId := range playlist.Songs {
		var audio models.Audio

		if err := initializers.DB.Where("id = ?", songId).First(&audio).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Missing audios"})
		}

		audios = append(audios, audio)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Audios fetched successfully",
		"playlistId": playlistId,
		"audios":     audios,
	})

}

func UpdatePlaylist(c *gin.Context) {

}

func DeletePlaylist(c *gin.Context) {

}

func GetPlaylistByProfile(c *gin.Context) {

}
