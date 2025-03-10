package handlers

import (
	"VoteGolang/database"
	"VoteGolang/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid president ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	db := database.GetDBInstance()
	var president models.President

	if err := db.First(&president, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "President not found"})
		return
	}

	var vote models.Vote
	result := db.Where("user_id = ? AND candidate_name = ?", userID, president.Name).First(&vote)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User has already voted for this president"})
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking vote"})
		return
	}

	vote = models.Vote{
		UserID:        userID.(uint),
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

	c.JSON(http.StatusOK, gin.H{"message": "Vote recorded successfully"})
}

func GetPresidentVoteCount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid president ID"})
		return
	}

	db := database.GetDBInstance()
	var president models.President

	if err := db.First(&president, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "President not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"votes": president.Votes})
}
