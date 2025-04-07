package usecases

import (
	"VoteGolang/internals/data"
	"VoteGolang/internals/repositories"
	"VoteGolang/pkg/domain"
	"fmt"
	"golang.org/x/crypto/bcrypt"
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
		return "", fmt.Errorf("user not found")
	}

	if !CheckPasswordHash(password, user.Password) {
		return "", fmt.Errorf("invalid credentials")
	}

	session := &domain.Session{
		ID:     generateSessionID(),
		UserID: user.ID,
	}

	token, err := a.TokenManager.Create(session, time.Now().Add(24*time.Hour).Unix())
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *AuthUseCase) Register(user *data.User) error {
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = hashedPassword

	err = a.UserRepo.Create(user)
	if err != nil {
		return fmt.Errorf("failed to register user: %v", err)
	}

	return nil
}

func generateSessionID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

func HashPassword(pw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(pw, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
	return err == nil
}
