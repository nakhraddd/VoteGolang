package auth_usecase

import (
	"VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/security"
	"fmt"
	"time"
)

// AuthUseCase handles authentication and authorization logic.
type AuthUseCase struct {
	UserRepo     domain.UserRepository
	TokenManager domain.TokenManager
}

func NewAuthUseCase(userRepo domain.UserRepository, tm domain.TokenManager) *AuthUseCase {
	return &AuthUseCase{
		UserRepo:     userRepo,
		TokenManager: tm,
	}
}

// Login authenticates a user and returns a JWT access tokens and refresh tokens.
func (a *AuthUseCase) Login(username, password string) (string, string, error) {
	u, err := a.UserRepo.GetByUsername(username)
	if err != nil {
		return "", "", fmt.Errorf("user not found")
	}

	if !security.CheckPasswordHash(password, u.Password) {
		return "", "", fmt.Errorf("invalid credentials")
	}

	accessToken, err := a.TokenManager.CreateAccessToken(u.ID, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := a.TokenManager.CreateRefreshToken(u.ID, 24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Register registers a new user with a hashed password.
func (a *AuthUseCase) Register(user *domain.User) error {

	if err := security.ValidatePassword(user.Password); err != nil {
		return err
	}
	hashedPassword, err := security.HashPassword(user.Password)
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
