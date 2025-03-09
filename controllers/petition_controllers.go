package controllers

import (
	"VoteGolang/database"
	"VoteGolang/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreatePetition(c *gin.Context) {
	var petition models.Petition
	if err := c.ShouldBindJSON(&petition); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Create(&petition)
	c.JSON(http.StatusCreated, gin.H{"message": "Petition created successfully"})
}
