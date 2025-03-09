package handlers

import (
	"VoteGolang/database"
	"VoteGolang/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetDeputyCandidates(c *gin.Context) {
	var candidates []models.Deputy
	db := database.GetDBInstance()

	if err := db.Find(&candidates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, candidates)
}
