package application

import (
	"context"

	"pray-together/internal/domains/invitation/domain"
)

// RejectInvitationRequest represents the request to reject an invitation
type RejectInvitationRequest struct {
	InvitationID uint64
	InviteeID    uint64
}

// RejectInvitationUseCase handles rejecting invitations
type RejectInvitationUseCase struct {
	invitationService *domain.Service
}

// NewRejectInvitationUseCase creates a new reject invitation use case
func NewRejectInvitationUseCase(invitationService *domain.Service) *RejectInvitationUseCase {
	return &RejectInvitationUseCase{
		invitationService: invitationService,
	}
}

// Execute rejects an invitation
func (u *RejectInvitationUseCase) Execute(ctx context.Context, req *RejectInvitationRequest) (*domain.InvitationInfo, error) {
	// Reject the invitation (returns error only)
	err := u.invitationService.RejectInvitation(ctx, req.InvitationID, req.InviteeID)
	if err != nil {
		return nil, err
	}

	// Get the invitation to return its info
	invitation, err := u.invitationService.GetInvitation(ctx, req.InvitationID)
	if err != nil {
		return nil, err
	}

	return invitation.ToInfo(), nil
}
