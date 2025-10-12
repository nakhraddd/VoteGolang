package auth_usecase

import (
	"VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/security"
	"context"
	"fmt"
	"time"
)

// AuthUseCase handles authentication and authorization logic.
type AuthUseCase struct {
	UserRepo      domain.UserRepository
	RoleRepo      domain.RoleRepository
	TokenManager  domain.TokenManager
	EmailVerifier domain.EmailVerifier
}

func NewAuthUseCase(userRepo domain.UserRepository, roleRepo domain.RoleRepository, tm domain.TokenManager, emailVerifier domain.EmailVerifier) *AuthUseCase {
	return &AuthUseCase{
		UserRepo:      userRepo,
		RoleRepo:      roleRepo,
		TokenManager:  tm,
		EmailVerifier: emailVerifier,
	}
}

// Login authenticates a user and returns a JWT access tokens and refresh tokens.
func (a *AuthUseCase) Login(username, password string) (string, string, error) {
	u, err := a.UserRepo.GetByUsername(username)
	if err != nil {
		return "", "", fmt.Errorf("user not found")
	}

	if !u.EmailVerified {
		return "", "", fmt.Errorf("email not verified")
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
func (a *AuthUseCase) Register(ctx context.Context, user *domain.User) (string, string, error) {
	if user.Username == "" {
		return "", "", fmt.Errorf("username is required")
	}
	if user.Email == "" {
		return "", "", fmt.Errorf("email is required")
	}
	if user.Password == "" {
		return "", "", fmt.Errorf("password is required")
	}
	if err := security.ValidatePassword(user.Password); err != nil {
		return "", "", err
	}
	hashedPassword, err := security.HashPassword(user.Password)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = hashedPassword

	role, err := a.RoleRepo.GetByName("member")
	if err != nil {
		return "", "", fmt.Errorf("default role not found: %v", err)
	}
	user.RoleID = role.ID

	err = a.UserRepo.Create(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to register user_repository: %v", err)
	}

	link, token, err := a.EmailVerifier.SendVerificationMail(ctx, user.Email)
	if err != nil {
		return "", "", err
	}

	return link, token, nil
}

func (a *AuthUseCase) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	// Verify refresh token
	userID, err := a.TokenManager.VerifyRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// (Optional) If you store refresh tokens in DB/Redis, check if this one is still valid
	// Example: if !a.UserRepo.IsRefreshTokenValid(userID, refreshToken) { return "", "", fmt.Errorf("revoked refresh token") }

	// Generate new tokens
	accessToken, err := a.TokenManager.CreateAccessToken(userID, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := a.TokenManager.CreateRefreshToken(userID, 24*time.Hour)
	if err != nil {
		return "", "", err
	}

	// (Optional) Save the new refresh token and revoke the old one in DB
	// a.UserRepo.RotateRefreshToken(userID, refreshToken, newRefreshToken)

	return accessToken, newRefreshToken, nil
}

func (a *AuthUseCase) VerifyEmail(ctx context.Context, token string) error {
	email, err := a.EmailVerifier.VerifyEmail(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired verification token: %v", err)
	}

	// нужно найти юзера по email
	user, err := a.UserRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	err = a.UserRepo.MarkEmailVerified(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to mark email verified: %v", err)
	}

	return nil
}
