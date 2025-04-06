package usecases

import (
	"VoteGolang/internals/repositories"
	domain2 "VoteGolang/pkg/domain"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AuthUseCase struct {
	UserRepo     repositories.UserRepository
	TokenManager domain2.TokenManager
}

func NewAuthUseCase(userRepo repositories.UserRepository, tm domain2.TokenManager) *AuthUseCase {
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

	session := &domain2.Session{
		ID:     generateSessionID(),
		UserID: user.ID,
	}
	token, err := a.TokenManager.Create(session, time.Now().Add(24*time.Hour).Unix())
	if err != nil {
		return "", err
	}

	return token, nil
}

func generateSessionID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func HashPassword(pw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), 14)
	return string(bytes), err
}

func CheckPasswordHash(pw, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
	return err == nil
}
