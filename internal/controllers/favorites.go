package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
* This method add songs to the favorite
 */
func AddToFavorite(c *gin.Context) {
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

	audioID := c.PostForm("audioId")
	if audioID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	var audio models.Audio
	if err := initializers.DB.Where("id = ?", audioID).First(&audio).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audio not found"})
		return
	}

	newFavorite := models.Favorite{UserID: userModel.ID, AudioID: audio.ID}
	if err := initializers.DB.FirstOrCreate(&newFavorite, newFavorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to favorites"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Added to favorites successfully"})
}

/*
* This method remove song from favorite list
 */
func DeleteFromFavorite(c *gin.Context) {
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

	audioID := c.PostForm("audioId")
	if audioID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	if err := initializers.DB.Where("user_id = ? AND audio_id = ?", userModel.ID, audioID).Delete(&models.Favorite{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from favorites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed from favorites successfully"})
}

/*
* This method fetches favorite audios of an user
 */
func GetAllFavorites(c *gin.Context) {
	// Assuming the user is authenticated and their ID is available
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

	// Create a slice to hold the favorite audios
	var favorites []models.Favorite
	if err := initializers.DB.Preload("Audio").Where("user_id = ?", userModel.ID).Find(&favorites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve favorites"})
		return
	}

	audioList := make([]map[string]interface{}, 0)
	for _, fav := range favorites {
		// Preload each owner for the audio separately if needed
		var owner models.User
		if err := initializers.DB.First(&owner, fav.Audio.Owner).Error; err != nil {
			// Handle the error, perhaps continue to the next favorite
			continue
		}

		audioInfo := map[string]interface{}{
			"id":       fav.Audio.ID,
			"title":    fav.Audio.Title,
			"about":    fav.Audio.About,
			"category": fav.Audio.Category,
			"file":     fav.Audio.AudioURL,
			"poster":   fav.Audio.CoverURL,
			"owner": map[string]interface{}{
				"id":   owner.ID,
				"name": owner.Name,
			},
		}
		audioList = append(audioList, audioInfo)
	}

	// Return the list of favorites with detailed audio information
	c.JSON(http.StatusOK, gin.H{"favorites": audioList})
}
