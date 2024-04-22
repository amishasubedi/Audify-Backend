package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
* CreatePlaylist creates a new playlist with a given title and visibility.
* It expects form data containing 'title' and 'visibility' fields.
* Requires user authentication.
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
* AddToPlaylist adds an existing audio track to a specified playlist.
* It expects form data with 'audioId' and 'playlistId'.
* Requires user authentication and ownership of the playlist.
 */
func AddToPlaylist(c *gin.Context) {
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
	playlistID := c.PostForm("playlistId")

	if audioID == "" || playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	var audio models.Audio
	if err := initializers.DB.Where("id = ?", audioID).First(&audio).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audio not found"})
		return
	}

	var playlist models.Playlist
	if err := initializers.DB.Where("id = ? AND owner_id = ?", playlistID, userModel.ID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found or not owned by user"})
		return
	}

	var existingAudios []models.Audio
	if err := initializers.DB.Model(&playlist).Association("Audios").Find(&existingAudios, "id = ?", audioID); err == nil && len(existingAudios) > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Audio already in playlist"})
		return
	}

	if err := initializers.DB.Model(&playlist).Association("Audios").Append(&audio); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add audio to playlist"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Audio added to playlist successfully"})
}

/*
* GetPlaylistDetailsByID retrieves details of a playlist by its ID.
* It uses a path parameter 'playlistId' to identify the playlist.
* No authentication required to view public playlists.
 */
func GetPlaylistDetailsByID(c *gin.Context) {
	playlistID := c.Param("playlistId")

	var playlist models.Playlist
	var owner models.User

	if err := initializers.DB.Preload("Audios").Where("id = ?", playlistID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	playlist.SetRandomCoverURL(initializers.DB)

	if err := initializers.DB.Model(&playlist).Update("cover_url", playlist.CoverURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update playlist cover"})
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
			"coverurl":   playlist.CoverURL,
			"owner_name": owner.Name,
			"owner_id":   owner.ID,
			"song_count": len(playlist.Audios),
		},
	})
}

// Query public playlists not owned by the current user that have at least one song
func GetPublicPlaylists(c *gin.Context) {
	var playlists []models.Playlist

	err := initializers.DB.
		Joins("JOIN playlist_audios ON playlist_audios.playlist_id = playlists.id").
		Joins("JOIN audios ON audios.id = playlist_audios.audio_id AND audios.deleted_at IS NULL").
		Where("playlists.visibility = 'public'").
		Group("playlists.id").
		Having("COUNT(audios.id) > 0").
		Find(&playlists).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public playlists", "details": err.Error()})
		return
	}

	response := make([]gin.H, 0)

	for _, playlist := range playlists {
		var owner models.User

		if err := initializers.DB.Where("id = ?", playlist.Owner).First(&owner).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch owner details", "details": err.Error()})
			return
		}

		var audioCount int64
		initializers.DB.Model(&models.Audio{}).
			Joins("JOIN playlist_audios ON playlist_audios.audio_id = audios.id").
			Where("playlist_audios.playlist_id = ?", playlist.ID).
			Count(&audioCount)

		playlistDetails := gin.H{
			"id":         playlist.ID,
			"title":      playlist.Title,
			"visibility": playlist.Visibility,
			"coverurl":   playlist.CoverURL,
			"owner_name": owner.Name,
			"song_count": audioCount,
		}

		response = append(response, playlistDetails)
	}

	c.JSON(http.StatusOK, gin.H{"playlists": response})
}

/*
* GetAudiosByPlaylist fetches all audio tracks associated with a specific playlist.
* It uses a path parameter 'playlistId' to identify the playlist.
* No authentication required to view public playlists.
 */
func GetAudiosByPlaylist(c *gin.Context) {
	playlistId := c.Param("playlistId")

	var playlist models.Playlist

	if err := initializers.DB.Preload("Audios").Where("id = ?", playlistId).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	audioList := make([]map[string]interface{}, len(playlist.Audios))
	for i, audio := range playlist.Audios {
		owner := models.User{}
		initializers.DB.First(&owner, audio.Owner)

		audioList[i] = map[string]interface{}{
			"id":       audio.ID,
			"title":    audio.Title,
			"about":    audio.About,
			"category": audio.Category,
			"file":     audio.AudioURL,
			"poster":   audio.CoverURL,
			"owner": map[string]interface{}{
				"name": owner.Name,
				"id":   owner.ID,
			},
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"playlist": gin.H{
			"id":    playlist.ID,
			"title": playlist.Title,
		},
		"audios": audioList,
	})
}

/*
* UpdatePlaylist updates the details of a specified playlist.
* It expects a JSON payload with 'id', 'title', and 'visibility'.
* Requires user authentication and ownership of the playlist.
 */
func UpdatePlaylist(c *gin.Context) {
	var payload struct {
		ID         uint   `json:"id"`
		Title      string `json:"title"`
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

	c.JSON(http.StatusOK, gin.H{"message": "Playlist updated successfully"})
}

/*
* RemoveFromPlaylist removes a single audio track from a specified playlist.
* It expects form data with 'audioId' and 'playlistId'.
* Requires user authentication and ownership of the playlist.
 */
func RemoveFromPlaylist(c *gin.Context) {
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
	playlistID := c.PostForm("playlistId")

	if audioID == "" || playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	var playlist models.Playlist
	if err := initializers.DB.Where("id = ? AND owner_id = ?", playlistID, userModel.ID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found or not owned by user"})
		return
	}

	var audio models.Audio
	if err := initializers.DB.First(&audio, audioID).Error; err == nil {
		if err := initializers.DB.Model(&playlist).Association("Audios").Delete(&audio); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove audio from playlist"})
			return
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audio not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Audio removed from playlist successfully"})
}

/*
* DeletePlaylist deletes an entire playlist by its ID.
* It uses a path parameter 'playlistId' to identify the playlist to be deleted.
* Requires user authentication and ownership of the playlist.
 */
func DeletePlaylist(c *gin.Context) {
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

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: playlistId"})
		return
	}

	var playlist models.Playlist
	if err := initializers.DB.Where("id = ? AND owner_id = ?", playlistID, userModel.ID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found or not owned by user"})
		return
	}

	if err := initializers.DB.Delete(&playlist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist deleted successfully"})
}
