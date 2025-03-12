package handlers

import (
	"VoteGolang/internals/app/backservices/database"
	"VoteGolang/internals/app/userservices/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetSessionDeputyCandidates(c *gin.Context) {
	var candidates []models.SessionDeputy
	db := database.GetDBInstance()

	if err := db.Find(&candidates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, candidates)
}

func VoteForSessionDeputy(c *gin.Context) {
	sessionDeputyName := c.Param("name")

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
	var sessionDeputy models.SessionDeputy

	if err := db.Where("name = ?", sessionDeputyName).First(&sessionDeputy).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session Deputy not found"})
		return
	}

	var vote models.Vote
	if err := db.Where("user_id = ? AND candidate_name = ?", uint(userIDUint), sessionDeputy.Name).First(&vote).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User has already voted for this session deputy"})
		return
	}

	vote = models.Vote{
		UserID:        uint(userIDUint),
		CandidateName: sessionDeputy.Name,
	}
	if err := db.Create(&vote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	sessionDeputy.Votes++
	if err := db.Save(&sessionDeputy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote recorded successfully", "votes": sessionDeputy.Votes})
}
