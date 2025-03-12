package handlers

import (
	"VoteGolang/internals/app/backservices/database"
	"VoteGolang/internals/app/userservices/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPetitions(c *gin.Context) {
	var petitions []models.Petition
	db := database.GetDBInstance()

	if err := db.Find(&petitions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, petitions)
}

func VotePetition(c *gin.Context) {
	petitionTitle := c.Param("title")

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

	voteType := c.Query("type")
	if voteType != "vote_in_favor" && voteType != "vote_against" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vote type. Use 'vote_in_favor' or 'vote_against'"})
		return
	}

	db := database.GetDBInstance()
	var petition models.Petition

	// Find petition by title
	if err := db.Where("title = ?", petitionTitle).First(&petition).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Petition not found"})
		return
	}

	// Check if user has already voted
	var existingVote models.Vote
	if err := db.Where("user_id = ? AND candidate_name = ?", uint(userIDUint), petition.Title).First(&existingVote).Error; err == nil {
		// User has already voted, prevent changing
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already voted and cannot change your vote"})
		return
	} else if err != gorm.ErrRecordNotFound {
		// Other DB error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking vote"})
		return
	}

	// User has not voted yet, record the vote
	newVote := models.Vote{
		UserID:        uint(userIDUint),
		CandidateName: petition.Title,
		VoteType:      voteType,
	}

	if err := db.Create(&newVote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	// Update vote count in the petitions table
	if voteType == "vote_in_favor" {
		petition.VotesInFavor++
	} else {
		petition.VotesAgainst++
	}

	// Save the updated petition vote count
	if err := db.Save(&petition).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Vote recorded successfully",
		"votes_in_favor": petition.VotesInFavor,
		"votes_against":  petition.VotesAgainst,
	})
}
