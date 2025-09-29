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
	UserID uint `json:"uid"`
	jwt.StandardClaims
}

func NewJwtToken(secret string) *JwtToken {
	return &JwtToken{Secret: []byte(secret)}
}

func (tk *JwtToken) CreateAccessToken(userID uint, ttl time.Duration) (string, error) {
	return tk.createToken(userID, ttl)
}

func (tk *JwtToken) CreateRefreshToken(userID uint, ttl time.Duration) (string, error) {
	return tk.createToken(userID, ttl)
}

func (tk *JwtToken) createToken(userID uint, ttl time.Duration) (string, error) {
	claims := JwtClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tk.Secret)
}

func (tk *JwtToken) VerifyAccessToken(ctx context.Context, tokenStr string) (uint, error) {
	return tk.verifyToken(tokenStr)
}

func (tk *JwtToken) VerifyRefreshToken(ctx context.Context, tokenStr string) (uint, error) {
	return tk.verifyToken(tokenStr)
}

func (tk *JwtToken) verifyToken(tokenStr string) (uint, error) {
	claims := &JwtClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return tk.Secret, nil
	})
	if err != nil {
		return 0, fmt.Errorf("invalid tokens: %w", err)
	}
	return claims.UserID, nil
}

func (j *JwtToken) GetSecret() []byte {
	return j.Secret
}
