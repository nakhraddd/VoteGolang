package handlers

import (
	"VoteGolang/database"
	"VoteGolang/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetSessionDeputyCandidates(c *gin.Context) {
	var candidates []models.SessionDeputy
	db := database.GetDBInstance()

	// Use GORM's Find method to get the candidates from the "session_deputy" table
	if err := db.Find(&candidates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the news as JSON
	c.JSON(http.StatusOK, candidates)
}
