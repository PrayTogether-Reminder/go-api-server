package room

import (
	"context"
	"fmt"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/domain"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/logger"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"gorm.io/gorm"
)

// RoomService handles room business logic
type RoomService struct {
	db                   *gorm.DB
	roomRepository       *RoomRepository
	memberRoomRepository *MemberRoomRepository
	memberValidator      MemberValidator
}

type MemberValidator interface {
	ValidateMemberExists(ctx context.Context, tx *gorm.DB, memberID int64) error
}

// NewRoomService creates a new RoomService instance
func NewRoomService(db *gorm.DB, roomRepository *RoomRepository, memberRoomRepository *MemberRoomRepository, memberValidation MemberValidator) *RoomService {
	return &RoomService{
		db:                   db,
		roomRepository:       roomRepository,
		memberRoomRepository: memberRoomRepository,
		memberValidator:      memberValidation,
	}
}

// FetchInfiniteScroll fetches rooms with infinite scroll pagination
// Java의 fetchRoomsInfiniteScroll 메서드와 동일한 로직
func (s *RoomService) FetchInfiniteScroll(ctx context.Context, memberID int64, request *InfiniteScrollRequest) (*InfiniteScrollResponse, error) {
	var response *InfiniteScrollResponse

	err := database.WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		roomInfos, err := s.fetchRoomInfosByMember(ctx, tx, memberID, request)
		if err != nil {
			return fmt.Errorf("방 목록 조회 실패: %w", err)
		}

		// Empty roomInfos case
		if len(roomInfos) == 0 {
			response = &InfiniteScrollResponse{make([]RoomInfo, 0)}
			return nil
		}

		// 2. Extract room IDs
		roomIDs := make([]int64, 0, len(roomInfos))
		for _, room := range roomInfos {
			roomIDs = append(roomIDs, room.ID)
		}

		// 3. Fetch member counts for roomInfos
		memberCounts, err := s.memberRoomRepository.FindMemberCountsByRoomIDs(ctx, tx, roomIDs)
		if err != nil {
			return fmt.Errorf("방별 멤버 수 조회 실패: %w", err)
		}

		// Convert to map for easy lookup
		countMap := make(map[int64]int64)
		for _, result := range memberCounts {
			countMap[result.RoomID] = result.MemberCount
		}

		// 4. Update member counts in room info
		for i := range roomInfos {
			if count, exists := countMap[roomInfos[i].ID]; exists {
				roomInfos[i].MemberCount = count
			}
		}

		// 5. Create response
		response = &InfiniteScrollResponse{roomInfos}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// fetchRoomInfosByMember fetches rooms for a member based on pagination
// Java의 fetchRoomInfosByMember 메서드와 동일한 로직
func (s *RoomService) fetchRoomInfosByMember(ctx context.Context, tx *gorm.DB, memberID int64, request *InfiniteScrollRequest) ([]RoomInfo, error) {
	// TODO: 전략 패턴으로 orderBy 및 dir에 따른 repository 메서드 차별화 구현 (time, name, memberCnt 등)

	// Initial request (first page)
	if request.After == DefaultAfter {
		roomInfos, err := s.roomRepository.FindRoomInfosByMemberIDInitial(ctx, tx, memberID, InfiniteScrollPageSize)
		if err != nil {
			return nil, fmt.Errorf("회원 방 조회 도중 오류 발생: %w", err)
		}
		return roomInfos, nil
	}

	// Parse cursor (after) as timestamp
	afterTime, err := time.Parse(time.RFC3339, request.After)
	if err != nil {
		return nil, fmt.Errorf("회원 방 조회 도중 오류 발생: %w", sharedError.ErrInvalidTimeFormat)
	}

	// Subsequent requests (after cursor)
	roomInfos, err := s.roomRepository.FindRoomInfosByMemberIDAfter(ctx, tx, memberID, afterTime, InfiniteScrollPageSize)
	if err != nil {
		return nil, fmt.Errorf("회원 방 조회 도중 오류 발생: %w", err)
	}
	return roomInfos, nil
}

// CreateRoom creates a new room and adds the member as OWNER
// Java의 createRoom 메서드와 동일한 로직
func (s *RoomService) CreateRoom(ctx context.Context, memberID int64, request *CreateRoomRequest) (*sharedHttp.MessageResponse, error) {
	var response *sharedHttp.MessageResponse

	err := database.WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		// 1. Create room
		room := model.NewRoom(request.Name, request.Description)
		if err := s.roomRepository.Create(ctx, tx, room); err != nil {
			return fmt.Errorf("방 생성 실패: %w", err)
		}

		// 2. Add member to room as OWNER
		memberRoom := model.NewRoomMember(memberID, room.ID, model.RoomRoleOwner, true)
		if err := s.memberRoomRepository.Create(ctx, tx, memberRoom); err != nil {
			return fmt.Errorf("방 멤버 추가 실패: %w", err)
		}

		response = &sharedHttp.MessageResponse{
			Message: "방 생성을 완료했습니다.",
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// ExitRoom removes a member from a room
// Java의 deleteRoom 메서드와 동일한 로직
func (s *RoomService) ExitRoom(ctx context.Context, memberID int64, roomID int64) (*sharedHttp.MessageResponse, error) {
	var response *sharedHttp.MessageResponse

	err := database.WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		if err := s.memberValidator.ValidateMemberExists(ctx, tx, memberID); err != nil {
			return err
		}

		deleted, err := s.memberRoomRepository.DeleteByMemberIDAndRoomID(ctx, tx, memberID, roomID)
		if err != nil {
			return fmt.Errorf("방-회원 관계 삭제 실패: %w", err)
		}

		if !deleted {
			return ErrMemberRoomNotFound
		}

		response = &sharedHttp.MessageResponse{
			Message: "방을 나갔습니다.",
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *RoomService) FetchMembersInRoom(ctx context.Context, memberID int64, roomID int64) (*FetchRoomMemberResponse, error) {
	var response *FetchRoomMemberResponse
	logger.FromContext(ctx).Info("[FetchMEmbersInRoom]", "start", time.Now())
	err := database.WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		// 1. Validate member exists in room
		err := s.ValidateMemberExistInRoom(ctx, tx, memberID, roomID)
		if err != nil {
			return err
		}

		// 2. Find room members with member information (JOIN query)
		roomMembers, err := s.memberRoomRepository.FindMemberRooms(ctx, tx, roomID)
		if err != nil {
			return fmt.Errorf("방 멤버 조회 실패: %w", err)
		}

		// 3. Convert to DTOs with phone number suffix
		var dtos []RoomMemberDto
		for _, rm := range roomMembers {
			phoneNumber, err := domain.NewPhoneNumber(rm.PhoneNumber)
			if err != nil {
				return err
			}
			dtos = append(dtos, RoomMemberDto{
				ID:                rm.MemberID,
				Name:              rm.Name,
				PhoneNumberSuffix: phoneNumber.GetSuffix(),
			})
		}

		response = &FetchRoomMemberResponse{
			RoomMembers: dtos,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// ValidateMemberExistInRoom checks if a member exists in a room
// Should be called within a transaction (tx is passed as parameter)
func (s *RoomService) ValidateMemberExistInRoom(ctx context.Context, tx *gorm.DB, memberID int64, roomID int64) error {
	existing, err := s.memberRoomRepository.IsExistMemberInRoom(ctx, tx, memberID, roomID)
	if err != nil {
		return fmt.Errorf("방에 회원이 있는지 검증하는 도중 오류 발생 : %w", err)
	}
	if !existing {
		return ErrMemberRoomNotFound
	}
	return nil
}
