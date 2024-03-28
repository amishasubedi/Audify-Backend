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

		newPlaylist.Audios = append(newPlaylist.Audios, audioModel)
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
			"songs":      newPlaylist.Audios,
		},
	})
}

/*
* This method fetches playlist details by id
 */
func GetPlaylistDetailsByID(c *gin.Context) {
	playlistID := c.Param("playlistId")

	var playlist models.Playlist
	var owner models.User

	if err := initializers.DB.Preload("Audios").Where("id = ?", playlistID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	if err := initializers.DB.Where("id = ?", playlist.Owner).First(&owner).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Owner not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"playlist": gin.H{
			"id":         playlist.ID,
			"title":      playlist.Title,
			"visibility": playlist.Visibility,
			"owner_name": owner.Name,
			"song_count": len(playlist.Audios),
		},
	})
}

/*
* This method fetches the audio details of playlist by id
 */
func GetAudiosByPlaylist(c *gin.Context) {
	playlistId := c.Param("playlistId")

	var playlist models.Playlist

	if err := initializers.DB.Preload("Audios").Where("id = ?", playlistId).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	audios := make([]interface{}, 0, len(playlist.Audios))
	for _, audio := range playlist.Audios {
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

/*
* This method updates the specified playlist details and optionally adds an audio item if provided
 */
func UpdatePlaylist(c *gin.Context) {
	var payload struct {
		ID         uint   `json:"id"`
		Title      string `json:"title"`
		Item       uint   `json:"item"`
		Visibility string `json:"visibility"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

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

	result := initializers.DB.Model(&models.Playlist{}).
		Where("id = ? AND owner_id = ?", payload.ID, userModel.ID).
		Updates(models.Playlist{Title: payload.Title, Visibility: payload.Visibility})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Updating playlist failed"})
		return
	}

	if payload.Item != 0 {
		var playlist models.Playlist
		if err := initializers.DB.Preload("Audios").Where("id = ?", payload.ID).First(&playlist).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
			return
		}

		var audio models.Audio
		if err := initializers.DB.Where("id = ?", payload.Item).First(&audio).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Audio not found"})
			return
		}

		err := initializers.DB.Model(&playlist).Association("Audios").Append(&audio)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add audio to playlist"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist updated successfully"})
}

/*
*
 */
func DeletePlaylist(c *gin.Context) {
	playlistID := c.Query("playlistId")
	resID := c.Query("resId")
	all := c.Query("all")

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

	if all == "yes" {
		result := initializers.DB.Where("id = ? AND owner_id = ?", playlistID, userModel.ID).Delete(&models.Playlist{})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete playlist"})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found or not owned by the user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Playlist deleted successfully"})

	} else if resID != "" {
		var playlist models.Playlist
		if err := initializers.DB.Preload("Songs").Where("id = ? AND owner_id = ?", playlistID, userModel.ID).First(&playlist).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
			return
		}

		var audio models.Audio
		if err := initializers.DB.Where("id = ?", resID).First(&audio).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Audio not found"})
			return
		}

		err := initializers.DB.Model(&playlist).Association("Songs").Delete(&audio)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove audio from playlist"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Audio removed from playlist successfully"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

/*
* This method gives all playlist made by user
 */
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
