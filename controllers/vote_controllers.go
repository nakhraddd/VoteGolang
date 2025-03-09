package controllers

import (
	"VoteGolang/database"
	"VoteGolang/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CastVote(c *gin.Context) {
	var vote models.Vote
	if err := c.ShouldBindJSON(&vote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Create(&vote)
	c.JSON(http.StatusCreated, gin.H{"message": "Vote cast successfully"})
}
