package auth

import (
	"context"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	"gorm.io/gorm"
)

// AuthUseCase orchestrates auth-specific application logic.
type AuthUseCase struct {
	db          *gorm.DB
	authService *AuthService
}

func NewAuthUseCase(db *gorm.DB, authService *AuthService) *AuthUseCase {
	return &AuthUseCase{
		db:          db,
		authService: authService,
	}
}

func (u *AuthUseCase) Login(ctx context.Context, request *LoginRequest) (*LoginResponse, error) {
	return u.authService.Login(ctx, u.db, request)
}

func (u *AuthUseCase) Signup(ctx context.Context, request *SignupRequest) error {
	return database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		return u.authService.Signup(ctx, tx, request)
	})
}
