package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"net/http"
	"strconv"

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

	var followersCount, followingsCount int64

	initializers.DB.Model(&models.User_Relations{}).Where("following_id = ?", profileId).Count(&followersCount)
	initializers.DB.Model(&models.User_Relations{}).Where("follower_id = ?", profileId).Count(&followingsCount)

	c.JSON(http.StatusOK, gin.H{
		"profile": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"avatar":     user.AvatarURL,
			"bio":        user.Bio,
			"followers":  followersCount,
			"followings": followingsCount,
		},
	})
}

// List all followers for a user
func ListFollowers(c *gin.Context) {
	userId := c.Param("userId")

	var relations []models.User_Relations
	initializers.DB.Preload("Follower").Where("following_id = ?", userId).Find(&relations)

	var followers []models.User
	for _, relation := range relations {
		followers = append(followers, relation.Follower)
	}

	c.JSON(http.StatusOK, gin.H{
		"followers": followers,
	})
}

// List all followings for a user
func ListFollowing(c *gin.Context) {
	userId := c.Param("userId")

	var relations []models.User_Relations
	initializers.DB.Preload("Following").Where("follower_id = ?", userId).Find(&relations)

	var followings []models.User
	for _, relation := range relations {
		followings = append(followings, relation.Following)
	}

	c.JSON(http.StatusOK, gin.H{
		"following": followings,
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

// followers - followings logic
func FollowUser(c *gin.Context) {
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

	followingIDStr := c.Param("followingId")
	followingID, err := strconv.ParseUint(followingIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if userModel.ID == uint(followingID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot follow oneself"})
		return
	}

	var existingRelation models.User_Relations
	if initializers.DB.Where("follower_id = ? AND following_id = ?", userModel.ID, followingID).First(&existingRelation).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already following this user"})
		return
	}

	newRelation := models.User_Relations{
		FollowerID:  userModel.ID,
		FollowingID: uint(followingID),
	}

	if err := initializers.DB.Create(&newRelation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Followed user successfully"})
}

func UnfollowUser(c *gin.Context) {
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

	followingIDStr := c.Param("followingId")
	followingID, err := strconv.ParseUint(followingIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if userModel.ID == uint(followingID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot unfollow oneself"})
		return
	}

	var existingRelation models.User_Relations
	result := initializers.DB.Where("follower_id = ? AND following_id = ?", userModel.ID, followingID).First(&existingRelation)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Relationship does not exist"})
		return
	}

	if err := initializers.DB.Delete(&existingRelation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unfollow user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unfollowed user successfully"})
}
