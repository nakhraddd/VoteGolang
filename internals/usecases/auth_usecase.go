package usecases

import (
	"VoteGolang/internals/domain"
	"VoteGolang/internals/repositories"
	"fmt"
	"time"
)

type AuthUseCase struct {
	UserRepo     repositories.UserRepository
	TokenManager domain.TokenManager
}

func NewAuthUseCase(userRepo repositories.UserRepository, tm domain.TokenManager) *AuthUseCase {
	return &AuthUseCase{
		UserRepo:     userRepo,
		TokenManager: tm,
	}
}

func (a *AuthUseCase) Login(username, password string) (string, error) {
	user, err := a.UserRepo.GetByUsername(username)
	if err != nil {
		return "", err
	}

	if !CheckPasswordHash(password, user.Password) {
		return "", fmt.Errorf("invalid credentials")
	}

	session := &domain.Session{
		ID:     generateSessionID(),
		UserID: uint32(user.ID),
	}
	token, err := a.TokenManager.Create(session, time.Now().Add(24*time.Hour).Unix())
	if err != nil {
		return "", err
	}

	return token, nil
}
