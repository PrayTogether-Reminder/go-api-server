package room

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/domain"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"gorm.io/gorm"
)

// RoomService handles room business logic
type RoomService struct {
	roomRepository       *RoomRepository
	memberRoomRepository *MemberRoomRepository
	memberValidator      MemberValidator
}

type MemberValidator interface {
	ValidateMemberExists(ctx context.Context, tx *gorm.DB, memberID int64) error
}

// NewRoomService creates a new RoomService instance
func NewRoomService(roomRepository *RoomRepository, memberRoomRepository *MemberRoomRepository, memberValidation MemberValidator) *RoomService {
	return &RoomService{
		roomRepository:       roomRepository,
		memberRoomRepository: memberRoomRepository,
		memberValidator:      memberValidation,
	}
}

// GetInfiniteScroll fetches rooms with infinite scroll pagination
// Java의 fetchRoomsInfiniteScroll 메서드와 동일한 로직
func (s *RoomService) GetInfiniteScroll(ctx context.Context, tx *gorm.DB, memberID int64, request *InfiniteScrollRequest) (*InfiniteScrollResponse, error) {
	roomInfos, err := s.getRoomInfosByMember(ctx, tx, memberID, request)
	if err != nil {
		return nil, fmt.Errorf("방 목록 조회 실패: %w", err)
	}

	if len(roomInfos) == 0 {
		return &InfiniteScrollResponse{Rooms: make([]RoomInfo, 0)}, nil
	}

	roomIDs := make([]int64, 0, len(roomInfos))
	for _, room := range roomInfos {
		roomIDs = append(roomIDs, room.ID)
	}

	memberCounts, err := s.memberRoomRepository.FindMemberCountsByRoomIDs(ctx, tx, roomIDs)
	if err != nil {
		return nil, fmt.Errorf("방별 멤버 수 조회 실패: %w", err)
	}

	countMap := make(map[int64]int64)
	for _, result := range memberCounts {
		countMap[result.RoomID] = result.MemberCount
	}

	for i := range roomInfos {
		if count, exists := countMap[roomInfos[i].ID]; exists {
			roomInfos[i].MemberCount = count
		}
	}

	return &InfiniteScrollResponse{Rooms: roomInfos}, nil
}

// getRoomInfosByMember fetches rooms for a member based on pagination
// Java의 getRoomInfosByMember 메서드와 동일한 로직
func (s *RoomService) getRoomInfosByMember(ctx context.Context, tx *gorm.DB, memberID int64, request *InfiniteScrollRequest) ([]RoomInfo, error) {
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
func (s *RoomService) CreateRoom(ctx context.Context, tx *gorm.DB, memberID int64, request *CreateRoomRequest) (*sharedHttp.MessageResponse, error) {
	room := model.NewRoom(request.Name, request.Description)
	if err := s.roomRepository.Create(ctx, tx, room); err != nil {
		return nil, fmt.Errorf("방 생성 실패: %w", err)
	}

	memberRoom := model.NewRoomMember(memberID, room.ID, model.RoomRoleOwner, true)
	if err := s.memberRoomRepository.Create(ctx, tx, memberRoom); err != nil {
		return nil, fmt.Errorf("방 멤버 추가 실패: %w", err)
	}

	return &sharedHttp.MessageResponse{
		Message: "방 생성을 완료했습니다.",
	}, nil
}

// ExitRoom removes a member from a room
// Java의 deleteRoom 메서드와 동일한 로직
func (s *RoomService) ExitRoom(ctx context.Context, tx *gorm.DB, memberID int64, roomID int64) (*sharedHttp.MessageResponse, error) {
	if err := s.memberValidator.ValidateMemberExists(ctx, tx, memberID); err != nil {
		return nil, err
	}

	deleted, err := s.memberRoomRepository.DeleteByMemberIDAndRoomID(ctx, tx, memberID, roomID)
	if err != nil {
		return nil, fmt.Errorf("방-회원 관계 삭제 실패: %w", err)
	}

	if !deleted {
		return nil, ErrMemberRoomNotFound
	}

	return &sharedHttp.MessageResponse{
		Message: "방을 나갔습니다.",
	}, nil
}

func (s *RoomService) GetMembersInRoom(ctx context.Context, tx *gorm.DB, memberID int64, roomID int64) (*FetchRoomMemberResponse, error) {

	if err := s.ValidateMemberExistInRoom(ctx, tx, memberID, roomID); err != nil {
		return nil, err
	}

	roomMembers, err := s.memberRoomRepository.FindMemberRooms(ctx, tx, roomID)
	if err != nil {
		return nil, fmt.Errorf("방 멤버 조회 실패: %w", err)
	}

	var dtos []RoomMemberDto
	for _, rm := range roomMembers {
		phoneNumber, err := domain.NewPhoneNumber(rm.PhoneNumber)
		if err != nil {
			return nil, err
		}
		dtos = append(dtos, RoomMemberDto{
			ID:                rm.MemberID,
			Name:              rm.Name,
			PhoneNumberSuffix: phoneNumber.GetSuffix(),
		})
	}

	return &FetchRoomMemberResponse{
		RoomMembers: dtos,
	}, nil
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

func (s *RoomService) GetRoomByID(ctx context.Context, tx *gorm.DB, roomID int64) (*model.Room, error) {
	room, err := s.roomRepository.FindByID(ctx, tx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("방을 찾을 수 없습니다: roomID=%d %w", roomID, err)
		}
		return nil, fmt.Errorf("방 조회 실패: %w", err)
	}
	return room, nil
}
