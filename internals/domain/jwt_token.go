package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type JwtToken struct {
	Secret []byte
}

type JwtClaims struct {
	UserID uint   `json:"uid"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.StandardClaims
}

func NewJwtToken(secret string) *JwtToken {
	return &JwtToken{Secret: []byte(secret)}
}

func (tk *JwtToken) CreateAccessToken(userID uint, ttl time.Duration) (string, error) {
	return tk.createToken(userID, ttl, "access")
}

func (tk *JwtToken) CreateRefreshToken(userID uint, ttl time.Duration) (string, error) {
	return tk.createToken(userID, ttl, "refresh")
}

func (tk *JwtToken) createToken(userID uint, ttl time.Duration, tokenType string) (string, error) {
	claims := JwtClaims{
		UserID: userID,
		Type:   tokenType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tk.Secret)
}

func (tk *JwtToken) VerifyAccessToken(ctx context.Context, tokenStr string) (uint, error) {
	return tk.verifyToken(tokenStr, "access")
}

func (tk *JwtToken) VerifyRefreshToken(ctx context.Context, tokenStr string) (uint, error) {
	return tk.verifyToken(tokenStr, "refresh")
}

func (tk *JwtToken) verifyToken(tokenStr string, expectedType string) (uint, error) {
	claims := &JwtClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return tk.Secret, nil
	})
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}
	if claims.Type != expectedType {
		return 0, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.Type)
	}
	return claims.UserID, nil
}

func (j *JwtToken) GetSecret() []byte {
	return j.Secret
}
