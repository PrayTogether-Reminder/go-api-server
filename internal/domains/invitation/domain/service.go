package domain

import (
	"context"
	"time"
)

// Service represents invitation domain service
type Service struct {
	repo          Repository
	validateRoom  func(ctx context.Context, roomID uint64) error
	joinRoom      func(ctx context.Context, roomID, memberID uint64) error
	getMemberName func(ctx context.Context, memberID uint64) (string, error)
}

// NewService creates a new invitation service
func NewService(
	repo Repository,
	validateRoom func(ctx context.Context, roomID uint64) error,
	joinRoom func(ctx context.Context, roomID, memberID uint64) error,
	getMemberName func(ctx context.Context, memberID uint64) (string, error),
) *Service {
	return &Service{
		repo:          repo,
		validateRoom:  validateRoom,
		joinRoom:      joinRoom,
		getMemberName: getMemberName,
	}
}

// CreateInvitation creates a new invitation
func (s *Service) CreateInvitation(ctx context.Context, roomID, inviterID, inviteeID uint64, message string) (*Invitation, error) {
	// Validate room exists
	if s.validateRoom != nil {
		if err := s.validateRoom(ctx, roomID); err != nil {
			return nil, err
		}
	}

	// Check if invitation already exists
	existing, err := s.repo.FindByRoomAndInvitee(ctx, roomID, inviteeID)
	if err != nil {
		return nil, err
	}

	if existing != nil && existing.IsPending() {
		return nil, ErrAlreadyInvited
	}

	// Get inviter name
	inviterName := ""
	if s.getMemberName != nil {
		name, err := s.getMemberName(ctx, inviterID)
		if err == nil {
			inviterName = name
		}
	}

	// Create invitation
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	invitation, err := NewInvitation(roomID, inviterID, inviteeID, message, expiresAt)
	if err != nil {
		return nil, err
	}

	// Set inviter name
	invitation.InviterName = inviterName

	// Save to repository
	if err := s.repo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

// AcceptInvitation accepts an invitation
func (s *Service) AcceptInvitation(ctx context.Context, invitationID, inviteeID uint64) error {
	// Get invitation
	invitation, err := s.repo.FindByID(ctx, invitationID)
	if err != nil {
		return err
	}

	if invitation == nil {
		return ErrInvitationNotFound
	}

	// Verify invitee
	if invitation.InviteeID != inviteeID {
		return ErrNotAuthorized
	}

	// Accept invitation
	if err := invitation.Accept(); err != nil {
		return err
	}

	// Update in repository
	if err := s.repo.Update(ctx, invitation); err != nil {
		return err
	}

	// Don't join room here - that's handled by the application layer
	// to match Java architecture where domain doesn't know about room joining

	return nil
}

// RejectInvitation rejects an invitation
func (s *Service) RejectInvitation(ctx context.Context, invitationID, inviteeID uint64) error {
	// Get invitation
	invitation, err := s.repo.FindByID(ctx, invitationID)
	if err != nil {
		return err
	}

	if invitation == nil {
		return ErrInvitationNotFound
	}

	// Verify invitee
	if invitation.InviteeID != inviteeID {
		return ErrNotAuthorized
	}

	// Reject invitation
	if err := invitation.Reject(); err != nil {
		return err
	}

	// Update in repository
	return s.repo.Update(ctx, invitation)
}

// GetPendingInvitations gets pending invitations for a member
func (s *Service) GetPendingInvitations(ctx context.Context, inviteeID uint64) ([]*Invitation, error) {
	return s.repo.FindPendingByInviteeID(ctx, inviteeID)
}

// GetInvitation gets an invitation by ID
func (s *Service) GetInvitation(ctx context.Context, invitationID uint64) (*Invitation, error) {
	invitation, err := s.repo.FindByID(ctx, invitationID)
	if err != nil {
		return nil, err
	}

	if invitation == nil {
		return nil, ErrInvitationNotFound
	}

	return invitation, nil
}

// CancelInvitation cancels an invitation
func (s *Service) CancelInvitation(ctx context.Context, invitationID, inviterID uint64) error {
	// Get invitation
	invitation, err := s.repo.FindByID(ctx, invitationID)
	if err != nil {
		return err
	}

	if invitation == nil {
		return ErrInvitationNotFound
	}

	// Verify inviter
	if invitation.InviterID != inviterID {
		return ErrNotAuthorized
	}

	// Only pending invitations can be cancelled
	if !invitation.IsPending() {
		return ErrAlreadyResponded
	}

	// Delete invitation
	return s.repo.Delete(ctx, invitationID)
}

// CleanupExpiredInvitations marks expired invitations
func (s *Service) CleanupExpiredInvitations(ctx context.Context) error {
	return s.repo.MarkExpiredInvitations(ctx, time.Now())
}

// SendInvitation creates and sends an invitation
func (s *Service) SendInvitation(ctx context.Context, roomID, inviterID, inviteeID uint64, message string, expiresAt time.Time) (*Invitation, error) {
	// Validate room exists
	if s.validateRoom != nil {
		if err := s.validateRoom(ctx, roomID); err != nil {
			return nil, err
		}
	}

	// Check if invitation already exists
	existing, err := s.repo.FindByRoomAndInvitee(ctx, roomID, inviteeID)
	if err != nil {
		return nil, err
	}

	if existing != nil && existing.IsPending() {
		return nil, ErrAlreadyInvited
	}

	// Get inviter name
	inviterName := ""
	if s.getMemberName != nil {
		name, err := s.getMemberName(ctx, inviterID)
		if err == nil {
			inviterName = name
		}
	}

	// Create invitation
	invitation, err := NewInvitation(roomID, inviterID, inviteeID, message, expiresAt)
	if err != nil {
		return nil, err
	}

	// Set inviter name
	invitation.InviterName = inviterName

	// Save to repository
	if err := s.repo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

// GetInvitationsByInvitee gets all invitations for an invitee
func (s *Service) GetInvitationsByInvitee(ctx context.Context, inviteeID uint64) ([]*Invitation, error) {
	return s.repo.FindByInviteeID(ctx, inviteeID)
}
