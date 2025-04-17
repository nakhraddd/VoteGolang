package auth_usecase

import (
	"VoteGolang/internals/data"
	"VoteGolang/internals/repositories/user_repository"
	"VoteGolang/internals/utils"
	"VoteGolang/pkg/domain"
	"fmt"
	"time"
)

type AuthUseCase struct {
	UserRepo     user_repository.UserRepository
	TokenManager domain.TokenManager
}

func NewAuthUseCase(userRepo user_repository.UserRepository, tm domain.TokenManager) *AuthUseCase {
	return &AuthUseCase{
		UserRepo:     userRepo,
		TokenManager: tm,
	}
}

func (a *AuthUseCase) Login(username, password string) (string, error) {
	user, err := a.UserRepo.GetByUsername(username)
	if err != nil {
		return "", fmt.Errorf("user_repository not found")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", fmt.Errorf("invalid credentials")
	}

	session := &domain.Session{
		ID:     utils.GenerateSessionID(),
		UserID: user.ID,
	}

	token, err := a.TokenManager.Create(session, time.Now().Add(24*time.Hour).Unix())
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *AuthUseCase) Register(user *data.User) error {
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = hashedPassword

	err = a.UserRepo.Create(user)
	if err != nil {
		return fmt.Errorf("failed to register user_repository: %v", err)
	}

	return nil
}
