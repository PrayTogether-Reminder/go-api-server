package domain

import (
	"context"
	"time"
)

// Repository interface for notification domain
type Repository interface {
	// Basic CRUD operations
	Create(ctx context.Context, notification *Notification) error
	FindByID(ctx context.Context, id uint64) (*Notification, error)
	Update(ctx context.Context, notification *Notification) error
	Delete(ctx context.Context, id uint64) error

	// Query operations
	FindByMemberID(ctx context.Context, memberID uint64, limit, offset int) ([]*Notification, error)
	FindUnreadByMemberID(ctx context.Context, memberID uint64) ([]*Notification, error)
	FindByType(ctx context.Context, memberID uint64, notifType NotificationType) ([]*Notification, error)
	FindPendingNotifications(ctx context.Context, limit int) ([]*Notification, error)

	// Count operations
	CountUnreadByMemberID(ctx context.Context, memberID uint64) (int, error)

	// Bulk operations
	MarkAllAsRead(ctx context.Context, memberID uint64) error
	DeleteOldNotifications(ctx context.Context, before time.Time) error
}
