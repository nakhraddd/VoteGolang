package handlers

import (
	"VoteGolang/database"
	"VoteGolang/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetPresidentCandidates(c *gin.Context) {
	var president []models.President
	db := database.GetDBInstance()

	// Use GORM's Find method to get the candidates from the "president" table
	if err := db.Find(&president).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the news as JSON
	c.JSON(http.StatusOK, president)
}
