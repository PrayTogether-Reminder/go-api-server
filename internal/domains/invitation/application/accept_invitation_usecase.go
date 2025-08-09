package application

import (
	"context"

	"pray-together/internal/domains/invitation/domain"
)

// AcceptInvitationRequest represents the request to accept an invitation
type AcceptInvitationRequest struct {
	InvitationID uint64
	InviteeID    uint64
}

// AcceptInvitationUseCase handles accepting invitations
type AcceptInvitationUseCase struct {
	invitationService *domain.Service
	joinRoom          func(ctx context.Context, roomID, memberID uint64) error
}

// NewAcceptInvitationUseCase creates a new accept invitation use case
func NewAcceptInvitationUseCase(
	invitationService *domain.Service,
	joinRoom func(ctx context.Context, roomID, memberID uint64) error,
) *AcceptInvitationUseCase {
	return &AcceptInvitationUseCase{
		invitationService: invitationService,
		joinRoom:          joinRoom,
	}
}

// Execute accepts an invitation
func (u *AcceptInvitationUseCase) Execute(ctx context.Context, req *AcceptInvitationRequest) (*domain.InvitationInfo, error) {
	// Get the invitation first to get room ID
	invitation, err := u.invitationService.GetInvitation(ctx, req.InvitationID)
	if err != nil {
		return nil, err
	}

	// Accept the invitation (matching Java: invitationService.accept)
	err = u.invitationService.AcceptInvitation(ctx, req.InvitationID, req.InviteeID)
	if err != nil {
		return nil, err
	}

	// Add member to room (matching Java: memberRoomService.addMemberToRoom with MEMBER role)
	if u.joinRoom != nil {
		if err := u.joinRoom(ctx, invitation.RoomID, req.InviteeID); err != nil {
			return nil, err
		}
	}

	return invitation.ToInfo(), nil
}
