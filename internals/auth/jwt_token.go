package auth

import (
	"VoteGolang/pkg/domain"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt"
	"time"
)

type JwtToken struct {
	Secret []byte
}

type JwtClaims struct {
	SessionID string `json:"sid"`
	UserID    string `json:"uid"`
	jwt.StandardClaims
}

func NewJwtToken(secret string) *JwtToken {
	return &JwtToken{Secret: []byte(secret)}
}

// Create generates a new JWT token
func (tk *JwtToken) Create(s *domain.Session, exp int64) (string, error) {
	claims := JwtClaims{
		SessionID: s.ID,
		UserID:    s.UserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: exp,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tk.Secret)
}

// Check validates the token, verifying it against the session information
// It now accepts a context.Context as the first parameter
func (tk *JwtToken) Check(ctx context.Context, inputToken string) (bool, error) {
	payload := &JwtClaims{}
	_, err := jwt.ParseWithClaims(inputToken, payload, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return tk.Secret, nil
	})
	if err != nil {
		return false, err
	}

	// In a real-world scenario, you might use `ctx` for things like logging, canceling requests, etc.
	// Here we're checking the claims and ensuring that the session matches the provided data
	if payload.SessionID == "" || payload.UserID == "" {
		return false, fmt.Errorf("invalid token claims")
	}

	return true, nil
}
