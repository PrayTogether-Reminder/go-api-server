package application

import (
	"context"
	"time"

	"pray-together/internal/domains/invitation/domain"
)

// SendInvitationRequest represents the request to send an invitation
type SendInvitationRequest struct {
	RoomID       uint64
	InviterID    uint64
	InviteeEmail string
	Message      string
	ExpiresAt    time.Time
}

// SendInvitationUseCase handles sending invitations
type SendInvitationUseCase struct {
	invitationService  *domain.Service
	getMemberByEmail   func(ctx context.Context, email string) (uint64, error)
	validateRoomAccess func(ctx context.Context, roomID, memberID uint64) error
}

// NewSendInvitationUseCase creates a new send invitation use case
func NewSendInvitationUseCase(
	invitationService *domain.Service,
	getMemberByEmail func(ctx context.Context, email string) (uint64, error),
	validateRoomAccess func(ctx context.Context, roomID, memberID uint64) error,
) *SendInvitationUseCase {
	return &SendInvitationUseCase{
		invitationService:  invitationService,
		getMemberByEmail:   getMemberByEmail,
		validateRoomAccess: validateRoomAccess,
	}
}

// Execute sends an invitation
func (u *SendInvitationUseCase) Execute(ctx context.Context, req *SendInvitationRequest) (*domain.InvitationInfo, error) {
	// Validate that inviter has access to the room (matching Java: validateMemberExistInRoom)
	if err := u.validateRoomAccess(ctx, req.RoomID, req.InviterID); err != nil {
		return nil, domain.ErrNotAuthorized
	}

	// Get invitee ID from email
	inviteeID, err := u.getMemberByEmail(ctx, req.InviteeEmail)
	if err != nil {
		return nil, err
	}

	// Validate that invitee is NOT already in the room (matching Java: validateMemberNotExistInRoom)
	if err := u.validateRoomAccess(ctx, req.RoomID, inviteeID); err == nil {
		// If no error, it means the member is already in the room
		return nil, domain.ErrAlreadyInvited
	}

	// Create and send invitation
	invitation, err := u.invitationService.SendInvitation(ctx, req.RoomID, req.InviterID, inviteeID, req.Message, req.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return invitation.ToInfo(), nil
}
