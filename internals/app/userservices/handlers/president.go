package handlers

import (
	"VoteGolang/internals/app/backservices/database"
	"VoteGolang/internals/app/userservices/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetPresidentCandidates(c *gin.Context) {
	var president []models.President
	db := database.GetDBInstance()

	if err := db.Find(&president).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, president)
}

func VoteForPresident(c *gin.Context) {
	presidentName := c.Param("name")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user ID"})
		return
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	db := database.GetDBInstance()
	var president models.President

	if err := db.Where("name = ?", presidentName).First(&president).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "President not found"})
		return
	}

	var vote models.Vote
	if err := db.Where("user_id = ? AND candidate_name = ?", uint(userIDUint), president.Name).First(&vote).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User has already voted for this president"})
		return
	}

	vote = models.Vote{
		UserID:        uint(userIDUint),
		CandidateName: president.Name,
	}
	if err := db.Create(&vote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	president.Votes++
	if err := db.Save(&president).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote recorded successfully", "votes": president.Votes})
}
