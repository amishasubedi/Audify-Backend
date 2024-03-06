package controllers

import (
	"backend/internal/initializers"
	"backend/internal/models"
	"backend/internal/utils"
	"fmt"
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

	fmt.Print("User Id", userModel.ID)

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
	// if user, exists := c.Get("user");
	// var audioURL, coverURL, audioPublicID, coverPublicID string

	// if !exists {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "unauthorized user"})
	// 	return
	// }

	// userModel, ok := user.(*models.User)
	// if !ok {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "User casting error"})
	// 	return
	// }

	// fmt.Print("User Id", userModel.ID)

	// title := c.PostForm("title")
	// about := c.PostForm("about")
	// category := c.PostForm("category")

	// if title == "" || about == "" || category == "" {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Should not be empty"})
	// 	return
	// }

}

/*
* List all audios for admin
 */
func GetAllAudios(c *gin.Context) {
	var audios []models.Audio

	if result := initializers.DB.Find(&audios); result.Error != nil {
		c.Error(result.Error)
		return
	}

	c.JSON(http.StatusOK, gin.H{"audios": audios})
}
