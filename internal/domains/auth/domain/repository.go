package domain

import (
	"context"
	"time"
)

// Repository interface for auth domain
type Repository interface {
	// RefreshToken operations
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	FindRefreshTokenByToken(ctx context.Context, token string) (*RefreshToken, error)
	FindRefreshTokensByMemberID(ctx context.Context, memberID uint64) ([]*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteRefreshTokensByMemberID(ctx context.Context, memberID uint64) error
	DeleteExpiredRefreshTokens(ctx context.Context) error

	// OTP operations
	CreateOTP(ctx context.Context, otp *OTP) error
	FindOTPByEmailAndCode(ctx context.Context, email, code, purpose string) (*OTP, error)
	FindLatestOTPByEmail(ctx context.Context, email, purpose string) (*OTP, error)
	UpdateOTP(ctx context.Context, otp *OTP) error
	DeleteOTP(ctx context.Context, id uint64) error
	DeleteExpiredOTPs(ctx context.Context) error
	DeleteOTPsByEmail(ctx context.Context, email string) error

	// Cleanup operations
	CleanupExpiredTokens(ctx context.Context, before time.Time) error
}
