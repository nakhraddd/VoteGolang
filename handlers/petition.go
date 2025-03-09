package handlers

import (
	"VoteGolang/database"
	"VoteGolang/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetPetitions(c *gin.Context) {
	var petitions []models.Petition
	db := database.GetDBInstance()

	// Use GORM's Find method to get the petitions from the "petition" table
	if err := db.Find(&petitions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the news as JSON
	c.JSON(http.StatusOK, petitions)
}
