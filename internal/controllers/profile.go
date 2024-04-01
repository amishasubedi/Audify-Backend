package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
* This method gives detail of the public profile
 */
func GetPublicProfile(c *gin.Context) {
	profileId := c.Param("userId")

	var user models.User
	result := initializers.DB.Where("id = ?", profileId).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"avatar":     user.AvatarURL,
			"followers":  len(user.Followers),
			"followings": len(user.Followings),
		},
	})
}

/*
* This  method fetches audios upload by you
 */
func GetPersonalUploads(c *gin.Context) {
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

	fmt.Println("User ID:", userModel.ID)

	var audios []models.Audio

	if err := initializers.DB.Where("owner = ?", userModel.ID).Find(&audios).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audio not found"})
		return
	}

	audioList := make([]map[string]interface{}, len(audios))
	for i, item := range audios {
		owner := models.User{}
		initializers.DB.First(&owner, item.Owner)

		audioList[i] = map[string]interface{}{
			"id":       item.ID,
			"title":    item.Title,
			"category": item.Category,
			"file":     item.AudioURL,
			"poster":   item.CoverURL,
			"owner": map[string]interface{}{
				"name": owner.Name,
				"id":   owner.ID,
			},
		}
	}

	c.JSON(http.StatusOK, gin.H{"audios": audioList})

}

/*
* This method fetches playlist and songs associated with it for personal profile
 */
func GetPersonalPlaylist(c *gin.Context) {
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

	fmt.Println("User ID:", userModel.ID)

	var playlists []models.Playlist

	if err := initializers.DB.Where("owner_id = ?", userModel.ID).Find(&playlists).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	playlistList := make([]map[string]interface{}, len(playlists))
	for i, item := range playlists {
		owner := models.User{}
		initializers.DB.First(&owner, item.Owner)

		playlistList[i] = map[string]interface{}{
			"id":    item.ID,
			"title": item.Title,
			"owner": map[string]interface{}{
				"name": owner.Name,
				"id":   owner.ID,
			},
		}
	}

	c.JSON(http.StatusOK, gin.H{"audios": playlistList})
}

func UpdateFollower(c *gin.Context) {

}
