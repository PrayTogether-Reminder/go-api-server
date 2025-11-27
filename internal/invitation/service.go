package invitation

import (
	"context"
	"errors"
	"fmt"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

type InvitationService struct {
	invitationRepository *InvitationRepository
}

func NewInvitationService(invitationRepository *InvitationRepository) *InvitationService {
	return &InvitationService{
		invitationRepository: invitationRepository,
	}
}

// CreatePending creates invitations for multiple invitees
// Filters out invitees who already have a PENDING invitation to the same room
func (s *InvitationService) CreatePending(ctx context.Context, db *gorm.DB, inviterName string, invitees []*model.Member, roomID int64) error {
	if len(invitees) == 0 {
		return nil
	}

	inviteeIDs := make([]int64, 0, len(invitees))
	for _, invitee := range invitees {
		inviteeIDs = append(inviteeIDs, invitee.ID)
	}

	// Find existing PENDING invitations
	existingInvitations, err := s.invitationRepository.FindByRoomIDAndStatusAndInviteeIDs(
		ctx, db, roomID, model.InvitationPending, inviteeIDs,
	)
	if err != nil {
		return fmt.Errorf("기도방 초대 응답 대기 조회 실패: %w", err)
	}

	// 초대를 받고 대기중인 회원 제외
	existingInviteeIDs := make(map[int64]bool)
	for _, invitation := range existingInvitations {
		existingInviteeIDs[invitation.InviteeID] = true
	}

	// Filter out invitees who already have PENDING invitations
	var newInvitations []*model.Invitation
	for _, invitee := range invitees {
		if existingInviteeIDs[invitee.ID] {
			continue
		}

		newInvitations = append(newInvitations, &model.Invitation{
			InviterName: inviterName,
			InviteeID:   invitee.ID,
			RoomID:      roomID,
			Status:      model.InvitationPending,
		})
	}

	// Create new invitations
	if len(newInvitations) > 0 {
		if err := s.invitationRepository.CreateAll(ctx, db, newInvitations); err != nil {
			return fmt.Errorf("기도방 초대 실패: %w", err)
		}
	}

	return nil
}

// FetchInvitationInfosByInviteeID retrieves pending invitation infos for a member
func (s *InvitationService) FetchInvitationInfosByInviteeID(ctx context.Context, db *gorm.DB, inviteeID int64) ([]InvitationInfo, error) {
	infos, err := s.invitationRepository.FindInfosByInviteeIDAndStatus(ctx, db, inviteeID, model.InvitationPending)
	if err != nil {
		return nil, fmt.Errorf("기도방 초대 목록 조회 실패: %w", err)
	}
	return infos, nil
}

// FetchByInviteeIDAndID retrieves an invitation by invitee ID and invitation ID
func (s *InvitationService) FetchByInviteeIDAndID(ctx context.Context, db *gorm.DB, inviteeID, invitationID int64) (*model.Invitation, error) {
	invitation, err := s.invitationRepository.FindByInviteeIDAndID(ctx, db, inviteeID, invitationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("기도방 초대장을 찾을 수 없습니다: %w", ErrInvitationNotFound)
		}
		return nil, fmt.Errorf("기도방 초대장 조회 실패: %w", err)
	}
	return invitation, nil
}

// Accept accepts an invitation
func (s *InvitationService) Accept(ctx context.Context, db *gorm.DB, invitation *model.Invitation) error {
	if err := invitation.Accept(); err != nil {
		return fmt.Errorf("이미 응답한 초대장입니다: %w", ErrAlreadyRespondedInvitation)
	}

	if err := s.invitationRepository.Update(ctx, db, invitation); err != nil {
		return fmt.Errorf("기도방 초대장 상태 업데이트 실패: %w", err)
	}

	return nil
}

// Reject rejects an invitation
func (s *InvitationService) Reject(ctx context.Context, db *gorm.DB, invitation *model.Invitation) error {
	if err := invitation.Reject(); err != nil {
		return fmt.Errorf("이미 응답한 초대장입니다: %w", ErrAlreadyRespondedInvitation)
	}

	if err := s.invitationRepository.Update(ctx, db, invitation); err != nil {
		return fmt.Errorf("기도방 초대장 상태 업데이트 실패: %w", err)
	}

	return nil
}
