package repositories

import (
	"VoteGolang/internals/app/connections"
	"VoteGolang/internals/data"
	"VoteGolang/internals/deliveries"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Id           string `json:"id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	UserFullName string `json:"userFullName"`
	BirthDate    string `json:"birthDate"`
	Address      string `json:"address"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func LoginHandler(c *gin.Context) {
	var loginRequest LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := connections.GetUserByUsername(loginRequest.Username)
	if err != nil || user.Password != loginRequest.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := deliveries.CreateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

func RegisterHandler(c *gin.Context) {
	var registerRequest RegisterRequest
	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newUser := data.User{
		ID:           registerRequest.Id,
		Username:     registerRequest.Username,
		Password:     registerRequest.Password,
		UserFullName: registerRequest.UserFullName,
		BirthDate:    registerRequest.BirthDate,
		Address:      registerRequest.Address,
	}

	if err := connections.CreateUser(&newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func getUserID(c *gin.Context) (uint, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, fmt.Errorf("Unauthorized")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return 0, fmt.Errorf("Invalid user ID format")
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid user ID format")
	}

	return uint(userIDUint), nil
}
