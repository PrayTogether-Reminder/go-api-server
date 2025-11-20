package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// AuthService handles low-level auth helpers such as password hashing/comparison.
type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// ComparePassword validates a plaintext password against a stored hash.
func (a *AuthService) ComparePassword(hashedPassword, plainPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInCorrectEmailPassword
		}
		return fmt.Errorf("비밀번호 검증 실패: %w", err)
	}

	return nil
}

// HashPassword hashes the provided plaintext password using bcrypt.
func (a *AuthService) HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("비밀번호 해싱 실패: %w", err)
	}

	return string(hashed), nil
}
