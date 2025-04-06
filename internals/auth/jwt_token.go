package auth

import (
	"VoteGolang/pkg/domain"
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

func (tk *JwtToken) Check(s *domain.Session, inputToken string) (bool, error) {
	payload := &JwtClaims{}
	_, err := jwt.ParseWithClaims(inputToken, payload, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return tk.Secret, nil
	})
	if err != nil || payload.SessionID != s.ID || payload.UserID != s.UserID {
		return false, err
	}
	return true, nil
}
