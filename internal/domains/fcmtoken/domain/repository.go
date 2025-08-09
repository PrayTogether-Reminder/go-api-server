package domain

import (
	"context"
)

// Repository interface for FCM token domain
type Repository interface {
	// Basic CRUD operations
	Create(ctx context.Context, token *FCMToken) error
	FindByToken(ctx context.Context, token string) (*FCMToken, error)
	Update(ctx context.Context, token *FCMToken) error
	Delete(ctx context.Context, id uint64) error
	DeleteByToken(ctx context.Context, memberID uint64, token string) error

	// Query operations
	FindByMemberID(ctx context.Context, memberID uint64) ([]*FCMToken, error)
	FindActiveByMemberID(ctx context.Context, memberID uint64) ([]*FCMToken, error)
	FindByDeviceID(ctx context.Context, memberID uint64, deviceID string) (*FCMToken, error)
	DeleteByDeviceID(ctx context.Context, memberID uint64, deviceID string) error

	// Bulk operations
	DeactivateByMemberID(ctx context.Context, memberID uint64) error
	DeleteByMemberID(ctx context.Context, memberID uint64) error // For Java-style registration
	DeleteStaleTokens(ctx context.Context, days int) error
}
