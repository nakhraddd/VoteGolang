package domain

import "context"

type EmailVerifier interface {
	SendVerificationMail(ctx context.Context, email string) (string, string, error)
	VerifyEmail(ctx context.Context, token string) (string, error) // возвращает email
}
