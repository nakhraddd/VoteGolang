package handlers

import (
	"VoteGolang/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Login handler to authenticate user and generate token
func Login(c *gin.Context) {
	// Get user credentials from request body (you can replace this with actual authentication logic)
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Authenticate the user (for example, check the username and password)
	// For simplicity, we assume a hardcoded user
	if loginData.Username == "admin" && loginData.Password == "password123" {
		// Generate JWT token
		token, err := utils.CreateToken("12345") // Replace "12345" with actual user ID
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Send the token as a response
		c.JSON(http.StatusOK, gin.H{"token": token})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}
