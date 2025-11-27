package invitation

import (
	"context"
	"fmt"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/logger"
	"gorm.io/gorm"
)

type InvitationUseCase struct {
	db                *gorm.DB
	invitationService *InvitationService
	roomService       *room.RoomService
	memberService     *member.MemberService
}

func NewInvitationUseCase(
	db *gorm.DB,
	invitationService *InvitationService,
	roomService *room.RoomService,
	memberService *member.MemberService,
) *InvitationUseCase {
	return &InvitationUseCase{
		db:                db,
		invitationService: invitationService,
		roomService:       roomService,
		memberService:     memberService,
	}
}

// InviteMembersToRoom invites members to a room (V2)
func (uc *InvitationUseCase) InviteMembersToRoom(ctx context.Context, inviterMemberID int64, request *InviteMembersRequest) (*sharedHttp.MessageResponse, error) {
	log := logger.FromContext(ctx)
	log.Info("회원 초대 시작", "inviterMemberID", inviterMemberID, "roomID", request.RoomID, "memberIDs", request.MemberIDs)
	memberIDs := uniqueInt64s(request.MemberIDs)

	var result *sharedHttp.MessageResponse
	err := database.WithTransaction(ctx, uc.db, func(tx *gorm.DB) error {
		// 1. 초대자 권한 검증
		if err := uc.roomService.ValidateMemberExistInRoom(ctx, tx, inviterMemberID, request.RoomID); err != nil {
			return err
		}

		// 2. 초대 대상 중 이미 방 멤버가 있는지 검증 (있으면 전체 요청 실패)
		if err := uc.roomService.ValidateMembersNotExistInRoom(ctx, tx, memberIDs, request.RoomID); err != nil {
			return err
		}

		// 3. 초대자 조회
		inviter, err := uc.memberService.GetByID(ctx, tx, inviterMemberID)
		if err != nil {
			return err
		}

		// 4. 초대 대상자 조회 (존재 여부 검증)
		invitees, err := uc.memberService.GetByIDs(ctx, tx, memberIDs)
		if err != nil {
			return err
		}
		if len(invitees) != len(memberIDs) {
			return fmt.Errorf("존재하지 않는 회원이 포함되어 있습니다: %w", member.ErrMemberNotFound)
		}

		// 5. 초대장 일괄 생성 (중복 PENDING 초대 필터링 포함)
		if err := uc.invitationService.CreatePending(ctx, tx, inviter.Name, invitees, request.RoomID); err != nil {
			return err
		}

		result = &sharedHttp.MessageResponse{Message: "초대를 완료했습니다."}
		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Info("회원 초대 완료", "inviterMemberID", inviterMemberID, "roomID", request.RoomID)
	return result, nil
}

// GetInvitationInfoScroll retrieves pending invitation infos for a member
func (uc *InvitationUseCase) GetInvitationInfoScroll(ctx context.Context, memberID int64) (*InvitationInfoScrollResponse, error) {
	log := logger.FromContext(ctx)
	log.Info("초대 목록 조회 시작")

	infos, err := uc.invitationService.FetchInvitationInfosByInviteeID(ctx, uc.db, memberID)
	if err != nil {
		return nil, err
	}

	log.Info("초대 목록 조회 완료")
	return &InvitationInfoScrollResponse{Invitations: infos}, nil
}

// UpdateInvitationStatus updates the status of an invitation (accept or reject)
//
//	func (uc *InvitationUseCase) UpdateInvitationStatus(ctx context.Context, memberID, invitationID int64, request *InvitationStatusUpdateRequest) (*sharedHttp.MessageResponse, error) {
//		log := logger.FromContext(ctx)
//		log.Info("초대 응답", "memberID", memberID, "invitationID", invitationID, "status", request.Status)
//
//		var result *sharedHttp.MessageResponse
//		err := database.WithTransaction(ctx, uc.db, func(tx *gorm.DB) error {
//			// 1. 초대장 조회 및 권한 검증
//			invitation, err := uc.invitationService.FetchByInviteeIDAndID(ctx, tx, memberID, invitationID)
//			if err != nil {
//				return err
//			}
//
//			// 2. 상태별 처리
//			status := model.InvitationStatus(request.Status)
//			switch status {
//			case model.InvitationAccepted:
//				// 수락
//				if err := uc.invitationService.Accept(ctx, tx, invitation); err != nil {
//					return err
//				}
//
//				// 방에 자동 가입
//				room, err := uc.roomService.GetRoomByID(ctx, tx, invitation.RoomID)
//				if err != nil {
//					return err
//				}
//
//				invitee, err := uc.memberService.GetByID(ctx, tx, invitation.InviteeID)
//				if err != nil {
//					return err
//				}
//
//				if err := uc.roomService.AddMemberToRoom(ctx, tx, invitee, room, model.RoomRoleMember); err != nil {
//					return err
//				}
//
//				result = &sharedHttp.MessageResponse{Message: "기도방 초대를 수락했습니다."}
//
//			case model.InvitationRejected:
//				// 거절
//				if err := uc.invitationService.Reject(ctx, tx, invitation); err != nil {
//					return err
//				}
//
//				result = &sharedHttp.MessageResponse{Message: "기도방 초대를 거절했습니다."}
//
//			default:
//				return fmt.Errorf("invalid invitation status: %s", request.Status)
//			}
//
//			return nil
//		})
//
//		if err != nil {
//			return nil, err
//		}
//
//		log.Info("초대 응답 완료", "memberID", memberID, "invitationID", invitationID, "status", request.Status)
//		return result, nil
//	}
func uniqueInt64s(values []int64) []int64 {
	if len(values) == 0 {
		return []int64{}
	}

	seen := make(map[int64]struct{}, len(values))
	unique := make([]int64, 0, len(values))
	for _, v := range values {
		if _, exists := seen[v]; exists {
			continue
		}
		seen[v] = struct{}{}
		unique = append(unique, v)
	}

	return unique
}
