package handlers

import (
	"VoteGolang/database"
	"VoteGolang/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"

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

func VoteForDeputy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deputy ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	db := database.GetDBInstance()
	var deputy models.Deputy

	if err := db.First(&deputy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deputy not found"})
		return
	}

	var vote models.Vote
	result := db.Where("user_id = ? AND candidate_name = ?", userID, deputy.Name).First(&vote) // Use candidate name
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User has already voted for this deputy"})
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking vote"})
		return
	}

	vote = models.Vote{
		UserID:        userID.(uint),
		CandidateName: deputy.Name, // Store candidate name
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

	c.JSON(http.StatusOK, gin.H{"message": "Vote recorded successfully"})
}

func GetDeputyVoteCount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deputy ID"})
		return
	}

	db := database.GetDBInstance()
	var deputy models.Deputy

	if err := db.First(&deputy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deputy not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"votes": deputy.Votes})
}
