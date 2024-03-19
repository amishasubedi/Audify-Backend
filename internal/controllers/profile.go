package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func GetPublicPlaylists(c *gin.Context) {
	userId := c.Param("userId")

	var playlists models.Playlist
	vis := "public"
	result := initializers.DB.
		Where("owner_id = ? AND visibility = ?", userId, vis).
		Order("created_at desc").
		Find(&playlists)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"playlists": true})
}

func GetPersonalUploads(c *gin.Context) {

}

func UpdateFollower(c *gin.Context) {

}
