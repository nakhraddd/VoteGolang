package handlers

import (
	"VoteGolang/database"
	"VoteGolang/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetGeneralNews(c *gin.Context) {
	var news []models.GeneralNews
	db := database.GetDBInstance()

	if err := db.Find(&news).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, news)
}
