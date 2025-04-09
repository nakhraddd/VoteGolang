package auth

import (
	"VoteGolang/pkg/domain"
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"time"
)

type JwtTokenManager struct {
	secretKey string
}

func NewJwtTokenManager(secretKey string) *JwtTokenManager {
	return &JwtTokenManager{
		secretKey: secretKey,
	}
}

func (t *JwtTokenManager) Create(session *domain.Session, expTime int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": session.UserID,
		"exp": expTime,
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.secretKey))
}

func (t *JwtTokenManager) Check(ctx context.Context, tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secretKey), nil
	})

	if err != nil {
		return false, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Printf("Claims: %v", claims)

		if exp, ok := claims["exp"].(float64); ok {
			log.Printf("Token Expiration: %v", exp)
			if time.Now().Unix() > int64(exp) {
				return false, fmt.Errorf("token expired")
			}
		}
		return true, nil
	}

	log.Println("Invalid token")
	return false, fmt.Errorf("invalid token")
}
