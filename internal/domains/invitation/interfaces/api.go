package interfaces

import (
	"context"

	"pray-together/internal/domains/invitation/domain"
)

// Module represents the invitation module interface
type Module struct {
	Service            *domain.Service
	GetMemberByEmail   func(ctx context.Context, email string) (uint64, error)
	ValidateRoomAccess func(ctx context.Context, roomID, memberID uint64) error
	JoinRoom           func(ctx context.Context, roomID, memberID uint64) error
	GetRoomInfo        func(ctx context.Context, roomID uint64) (name string, description string, err error)
	GetMemberInfo      func(ctx context.Context, memberID uint64) (string, error)
}

// NewModule creates a new invitation module
func NewModule(
	repo domain.Repository,
	getMemberByEmail func(ctx context.Context, email string) (uint64, error),
	validateRoomAccess func(ctx context.Context, roomID, memberID uint64) error,
	joinRoom func(ctx context.Context, roomID, memberID uint64) error,
	getRoomInfo func(ctx context.Context, roomID uint64) (name string, description string, err error),
	getMemberInfo func(ctx context.Context, memberID uint64) (string, error),
) *Module {
	// Create validateRoom function for domain service
	validateRoom := func(ctx context.Context, roomID uint64) error {
		// For now, just return nil to accept all rooms
		// Could use validateRoomAccess if needed
		return nil
	}

	// Create getMemberName function for domain service
	getMemberName := func(ctx context.Context, memberID uint64) (string, error) {
		if getMemberInfo != nil {
			return getMemberInfo(ctx, memberID)
		}
		return "Member", nil
	}

	service := domain.NewService(repo, validateRoom, joinRoom, getMemberName)

	return &Module{
		Service:            service,
		GetMemberByEmail:   getMemberByEmail,
		ValidateRoomAccess: validateRoomAccess,
		JoinRoom:           joinRoom,
		GetRoomInfo:        getRoomInfo,
		GetMemberInfo:      getMemberInfo,
	}
}
