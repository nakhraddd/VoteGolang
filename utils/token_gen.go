package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Retrieve the secret key from the environment variable
var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func init() {
	// Ensure the secret key is set
	if len(secretKey) == 0 {
		fmt.Println("JWT_SECRET_KEY environment variable is not set!")
	}
}

// CreateToken generates a JWT token for a given user ID
func CreateToken(userID string) (string, error) {
	// Set token claims
	claims := jwt.MapClaims{
		"user_id": userID,                                // User ID stored in the claims
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Token expiration time (24 hours)
		"iat":     time.Now().Unix(),                     // Issued at time
	}

	// Create the token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	return tokenString, nil
}
