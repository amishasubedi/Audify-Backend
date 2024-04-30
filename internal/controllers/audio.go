package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
* This method uploads new music
 */
func CreateAudio(c *gin.Context) {
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
	about := c.PostForm("about")
	category := c.PostForm("category")

	if title == "" || about == "" || category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	var audioURL, coverURL, audioPublicID, coverPublicID string
	audioFile, err := c.FormFile("audioFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file is missing"})
		return
	}

	audioFilePath := audioFile.Filename
	audio, err := audioFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open audio file"})
		return
	}
	defer audio.Close()

	audioURL, audioPublicID, err = utils.UploadToCloudinary(audio, audioFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload audio file"})
		return
	}

	coverFile, coverErr := c.FormFile("coverFile")
	if coverErr == nil {
		coverFilePath := coverFile.Filename
		cover, err := coverFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open cover file"})
			return
		}
		defer cover.Close()

		coverURL, coverPublicID, err = utils.UploadToCloudinary(cover, coverFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload cover file"})
			return
		}
	}

	newAudio := models.Audio{
		Title:         title,
		About:         about,
		Category:      category,
		Owner:         userModel.ID,
		AudioURL:      audioURL,
		CoverURL:      coverURL,
		AudioPublicID: audioPublicID,
		CoverPublicID: coverPublicID,
	}

	if err := initializers.DB.Create(&newAudio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save audio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Audio created successfully",
		"audio":   newAudio,
	})
}

/*
*
 */
func UpdateAudio(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unauthorized user"})
		return
	}

	userModel, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User casting error"})
		return
	}

	audioId := c.Param("audioId")

	updates := make(map[string]interface{})
	for _, field := range []string{"name", "about", "category"} {
		if value := c.PostForm(field); value != "" {
			updates[field] = value
		}
	}

	var audio models.Audio
	if err := initializers.DB.First(&audio, "id = ? AND owner = ?", audioId, userModel.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audio not found or user not authorized to update this audio"})
		return
	}

	coverFile, _ := c.FormFile("coverFile")
	if coverFile != nil {
		if audio.CoverPublicID != "" {
			if err := utils.DestroyImage(audio.CoverPublicID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove existing cover image"})
				return
			}
		}

		file, err := coverFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open cover file"})
			return
		}
		defer file.Close()

		coverURL, coverPublicID, err := utils.UploadToCloudinary(file, coverFile.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload new cover image"})
			return
		}
		updates["cover_url"] = coverURL
		updates["cover_public_id"] = coverPublicID
	}

	if err := initializers.DB.Model(&audio).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update audio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Audio updated successfully",
		"data":    audio,
	})
}

/*
* List all audios for admin
 */
func GetLatestAudios(c *gin.Context) {
	var audios []models.Audio

	if result := initializers.DB.Find(&audios); result.Error != nil {
		c.Error(result.Error)
		return
	}

	c.JSON(http.StatusOK, gin.H{"audios": audios})
}

/*
* Get latest uploads for unauthorized user
 */
func GetLatestUploads(c *gin.Context) {
	var audios []models.Audio

	if err := initializers.DB.Order("created_at desc").Limit(12).Find(&audios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query latest uploads"})
		return
	}

	audioList := make([]map[string]interface{}, len(audios))
	for i, item := range audios {
		owner := models.User{}
		initializers.DB.First(&owner, item.Owner)

		audioList[i] = map[string]interface{}{
			"id":       item.ID,
			"title":    item.Title,
			"about":    item.About,
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
* fetch random 8 songs from the database
 */
func GetSuggestionsList(c *gin.Context) {
	var audios []models.Audio

	if err := initializers.DB.Order("RANDOM()").Limit(3).Find(&audios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query random songs"})
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

func FilterByMood(c *gin.Context) {
	category := c.Query("category")

	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category parameter is required"})
		return
	}

	var audios []models.Audio

	if err := initializers.DB.Order("created_at desc").Where("category = ?", category).Find(&audios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query audios by category"})
		return
	}

	if len(audios) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No audios found for the specified category.",
			"audios":  []interface{}{},
		})
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

func GetUploadsById(c *gin.Context) {
	userId := c.Param("userId")

	var audios []models.Audio

	if err := initializers.DB.Where("owner = ?", userId).Find(&audios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query user's uploads"})
		return
	}

	if len(audios) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No uploads found for the specified user.",
			"audios":  []interface{}{},
		})
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

// search audios or artist
func GeneralSearch(c *gin.Context) {
	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query cannot be empty"})
		return
	}

	var artists []models.User
	artistErr := initializers.DB.
		Where("name LIKE ?", "%"+query+"%").
		Find(&artists).Error

	if artistErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed in artists", "details": artistErr.Error()})
		return
	}

	results := gin.H{
		"artists": artists,
	}

	c.JSON(http.StatusOK, results)
}
