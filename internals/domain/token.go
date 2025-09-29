package domain

import "context"
import "time"

type TokenManager interface {
	CreateAccessToken(userID uint, ttl time.Duration) (string, error)
	CreateRefreshToken(userID uint, ttl time.Duration) (string, error)
	VerifyAccessToken(ctx context.Context, token string) (uint, error)  // return userID
	VerifyRefreshToken(ctx context.Context, token string) (uint, error) // return userID
	GetSecret() []byte
}
