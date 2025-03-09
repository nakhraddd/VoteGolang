package middleware

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
)

// Retrieve the secret key from the environment variable
var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func init() {
	// Ensure the secret key is set
	if len(secretKey) == 0 {
		fmt.Println("JWT_SECRET_KEY environment variable is not set!")
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token here
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm used for signing the token (e.g., HMAC, RSA)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil // Use the secret key from the environment
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// You can access the claims if needed (e.g., user info or roles)
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			userID := claims["user_id"]
			fmt.Println("Authenticated user ID:", userID)
		}

		c.Next()
	}
}
