package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"fmt"
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

		newPlaylist.Songs = append(newPlaylist.Songs, audioModel)
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

	audios := make([]interface{}, 0, len(playlist.Songs))
	for _, audio := range playlist.Songs {
		audios = append(audios, gin.H{
			"id":              audio.ID,
			"title":           audio.Title,
			"about":           audio.About,
			"owner":           audio.Owner,
			"audio_url":       audio.AudioURL,
			"audio_public_id": audio.AudioPublicID,
			"cover_url":       audio.CoverURL,
			"cover_public_id": audio.CoverPublicID,
			"likes":           audio.Likes,
			"category":        audio.Category,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"playlist": gin.H{
			"id":    playlist.ID,
			"title": playlist.Title,
		},
		"audios": audios,
	})

}

func UpdatePlaylist(c *gin.Context) {

}

func DeletePlaylist(c *gin.Context) {

}

func GetPlaylistByProfile(c *gin.Context) {
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

	fmt.Println("User Id", userModel.ID)

	var playlists []models.Playlist
	if err := initializers.DB.Where("owner_id = ?", userModel.ID).Find(&playlists).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to fetch playlists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
}
