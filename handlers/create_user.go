package handlers

import (
	"VoteGolang/database"
	"VoteGolang/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateUser(c *gin.Context) {
	var user models.User

	// Bind the request body to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that user data is not empty
	if user.Username == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	// Save the user to the database
	db := database.GetDBInstance()
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Return the created user
	c.JSON(http.StatusOK, gin.H{"message": "User created successfully", "user": user})
}

func GetUser(c *gin.Context) {
	id := c.Param("id") // Get the user ID from the URL

	// Find the user in the database by ID
	var user models.User
	db := database.GetDBInstance()
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return the user data
	c.JSON(http.StatusOK, user)
}
