package handlers

import (
	"VoteGolang/database"
	"VoteGolang/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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

func VoteForDeputy(c *gin.Context) {
	deputyName := c.Param("name")

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
	var deputy models.Deputy

	if err := db.Where("name = ?", deputyName).First(&deputy).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deputy not found"})
		return
	}

	var vote models.Vote
	if err := db.Where("user_id = ? AND candidate_name = ?", uint(userIDUint), deputy.Name).First(&vote).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User has already voted for this deputy"})
		return
	}

	vote = models.Vote{
		UserID:        uint(userIDUint),
		CandidateName: deputy.Name,
	}
	if err := db.Create(&vote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	deputy.Votes++
	if err := db.Save(&deputy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote recorded successfully", "votes": deputy.Votes})
}
