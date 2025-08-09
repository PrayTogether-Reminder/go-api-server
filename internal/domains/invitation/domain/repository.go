package domain

import (
	"context"
	"time"
)

// Repository interface for invitation domain
type Repository interface {
	// Basic CRUD operations
	Create(ctx context.Context, invitation *Invitation) error
	FindByID(ctx context.Context, id uint64) (*Invitation, error)
	Update(ctx context.Context, invitation *Invitation) error
	Delete(ctx context.Context, id uint64) error

	// Query operations
	FindByInviteeID(ctx context.Context, inviteeID uint64) ([]*Invitation, error)
	FindByRoomID(ctx context.Context, roomID uint64) ([]*Invitation, error)
	FindPendingByInviteeID(ctx context.Context, inviteeID uint64) ([]*Invitation, error)
	FindByRoomAndInvitee(ctx context.Context, roomID, inviteeID uint64) (*Invitation, error)

	// Cleanup operations
	MarkExpiredInvitations(ctx context.Context, before time.Time) error
}
