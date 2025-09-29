package security

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), 14)
	return string(bytes), err
}

func CheckPasswordHash(pw, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
	return err == nil
}

func ValidatePassword(pw string) error {
	var (
		hasUpperCase bool
		hasLowerCase bool
		hasDigit     bool
		hasSpecial   bool
	)

	for _, char := range pw {
		switch {
		case unicode.IsUpper(char):
			hasUpperCase = true
		case unicode.IsLower(char):
			hasLowerCase = true
		case unicode.IsDigit(char):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if len(pw) < 8 {
		return errors.New("password is too short, it should be at least 8 characters long")
	}
	if !hasUpperCase {
		return errors.New("password should contain at least one uppercase letter")
	}
	if !hasLowerCase {
		return errors.New("password should contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password should contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password should contain at least one special character")
	}

	return nil
}
