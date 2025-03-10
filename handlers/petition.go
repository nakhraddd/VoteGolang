package handlers

import (
	"VoteGolang/database"
	"VoteGolang/models"
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
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid petition ID"})
		return
	}

	voteType := c.Query("type")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	db := database.GetDBInstance()
	var petition models.Petition

	if err := db.First(&petition, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Petition not found"})
		return
	}

	var vote models.PetitionVote
	result := db.Where("user_id = ? AND petition_title = ?", userID, petition.Title).First(&vote) // Use petition title
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User has already voted for this petition"})
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking vote"})
		return
	}

	vote = models.PetitionVote{
		UserID:        userID.(uint),
		PetitionTitle: petition.Title, // Store petition title
		VoteType:      voteType,
	}

	if err := db.Create(&vote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	switch voteType {
	case "in_favor":
		petition.VotesInFavor++
	case "against":
		petition.VotesAgainst++
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vote type. Use 'in_favor' or 'against' "})
		return
	}

	if err := db.Save(&petition).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote recorded successfully"})
}

func GetPetitionVotes(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid petition ID"})
		return
	}

	db := database.GetDBInstance()
	var petition models.Petition

	if err := db.First(&petition, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Petition not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"votes_in_favor": petition.VotesInFavor, "votes_against": petition.VotesAgainst})
}
