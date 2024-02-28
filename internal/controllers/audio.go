package controllers

import (
	"backend/internal/models"

	"github.com/gin-gonic/gin"
)

/*
* This method uploads new music
 */
func CreateAudio(c *gin.Context) {
	var newAudio models.Audio

	if err := c.BindJSON(&newAudio); err != nil {
		c.Error(err)
		return
	}

}

/*
*
 */
func UpdateAudio(c *gin.Context) {

}
