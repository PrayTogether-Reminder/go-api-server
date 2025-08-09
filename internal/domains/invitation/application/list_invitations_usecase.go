package application

import (
	"context"

	"pray-together/internal/domains/invitation/domain"
)

// ListInvitationsRequest represents the request to list invitations
type ListInvitationsRequest struct {
	InviteeID      uint64
	Status         string
	IncludeExpired bool
}

// ListInvitationsUseCase handles listing invitations
type ListInvitationsUseCase struct {
	invitationService *domain.Service
	getRoomInfo       func(ctx context.Context, roomID uint64) (name string, description string, err error)
	getMemberInfo     func(ctx context.Context, memberID uint64) (string, error)
}

// NewListInvitationsUseCase creates a new list invitations use case
func NewListInvitationsUseCase(
	invitationService *domain.Service,
	getRoomInfo func(ctx context.Context, roomID uint64) (name string, description string, err error),
	getMemberInfo func(ctx context.Context, memberID uint64) (string, error),
) *ListInvitationsUseCase {
	return &ListInvitationsUseCase{
		invitationService: invitationService,
		getRoomInfo:       getRoomInfo,
		getMemberInfo:     getMemberInfo,
	}
}

// Execute lists invitations for a user
func (u *ListInvitationsUseCase) Execute(ctx context.Context, req *ListInvitationsRequest) ([]*domain.InvitationInfo, error) {
	// Get invitations from service
	invitations, err := u.invitationService.GetInvitationsByInvitee(ctx, req.InviteeID)
	if err != nil {
		return nil, err
	}

	// Filter by status if specified
	var filtered []*domain.Invitation
	for _, inv := range invitations {
		// Skip expired invitations unless explicitly requested
		if !req.IncludeExpired && inv.IsExpired() {
			continue
		}

		// Filter by status if specified
		if req.Status != "" && string(inv.Status) != req.Status {
			continue
		}

		filtered = append(filtered, inv)
	}

	// Convert to info objects with additional data
	result := make([]*domain.InvitationInfo, 0, len(filtered))
	for _, inv := range filtered {
		info := inv.ToInfo()

		// Get room name and description
		if roomName, roomDesc, err := u.getRoomInfo(ctx, inv.RoomID); err == nil {
			info.RoomName = roomName
			info.RoomDescription = roomDesc
		}

		// Get inviter name (already set in invitation)
		// But override if we can get fresher data
		if inviterName, err := u.getMemberInfo(ctx, inv.InviterID); err == nil {
			info.InviterName = inviterName
		}

		result = append(result, info)
	}

	return result, nil
}
