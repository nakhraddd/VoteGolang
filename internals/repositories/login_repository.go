package repositories

import (
	"VoteGolang/internals/data"
	"VoteGolang/pkg/domain"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"time"
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

var userRepo UserRepository
var tokenManager domain.TokenManager

func LoginHandler(c *gin.Context) {
	var loginRequest LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := userRepo.GetByUsername(loginRequest.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	session := &domain.Session{
		UserID: user.ID,
		Expiry: time.Now().Add(24 * time.Hour).Unix(),
	}

	token, err := tokenManager.Create(session, session.Expiry)
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	newUser := data.User{
		ID:           registerRequest.Id,
		Username:     registerRequest.Username,
		Password:     string(hashedPassword),
		UserFullName: registerRequest.UserFullName,
		BirthDate:    registerRequest.BirthDate,
		Address:      registerRequest.Address,
	}

	if err := userRepo.Create(&newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func getUserID(c *gin.Context) (uint, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, fmt.Errorf("unauthorized")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return 0, fmt.Errorf("invalid user ID format")
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format")
	}

	return uint(userIDUint), nil
}
