package room

import (
	"context"
	"fmt"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/logger"
	"gorm.io/gorm"
)

// RoomService handles room business logic
type RoomService struct {
	db                   *gorm.DB
	roomRepository       *RoomRepository
	memberRoomRepository *MemberRoomRepository
}

// NewRoomService creates a new RoomService instance
func NewRoomService(db *gorm.DB, roomRepository *RoomRepository, memberRoomRepository *MemberRoomRepository) *RoomService {
	return &RoomService{
		db:                   db,
		roomRepository:       roomRepository,
		memberRoomRepository: memberRoomRepository,
	}
}

// FetchInfiniteScroll fetches rooms with infinite scroll pagination
// Java의 fetchRoomsInfiniteScroll 메서드와 동일한 로직
func (s *RoomService) FetchInfiniteScroll(ctx context.Context, memberID uint32, request *InfiniteScrollRequest) (*InfiniteScrollResponse, error) {
	log := logger.FromContext(ctx)
	var response *InfiniteScrollResponse

	err := database.WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		roomInfos, err := s.fetchRoomInfosByMember(ctx, tx, memberID, request)
		if err != nil {
			log.Error("회원의 방 목록 조회 실패", "memberID", memberID, "error", err)
			return fmt.Errorf("방 목록 조회 실패: %w", err)
		}

		// Empty roomInfos case
		if len(roomInfos) == 0 {
			response = NewInfiniteScrollResponse([]RoomInfo{})
			return nil
		}

		// 2. Extract room IDs
		roomIDs := make([]uint32, 0, len(roomInfos))
		for _, room := range roomInfos {
			roomIDs = append(roomIDs, room.ID)
		}

		// 3. Fetch member counts for roomInfos
		memberCounts, err := s.memberRoomRepository.FindMemberCountsByRoomIDs(ctx, tx, roomIDs)
		if err != nil {
			log.Error("방별 멤버 수 조회 실패", "roomIDs", roomIDs, "error", err)
			return fmt.Errorf("방별 멤버 수 조회 실패: %w", err)
		}

		// Convert to map for easy lookup
		countMap := make(map[uint32]int)
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
		response = NewInfiniteScrollResponse(roomInfos)
		log.Info("방 목록 조회 성공", "memberID", memberID, "roomCount", len(roomInfos))
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// fetchRoomInfosByMember fetches rooms for a member based on pagination
// Java의 fetchRoomInfosByMember 메서드와 동일한 로직
func (s *RoomService) fetchRoomInfosByMember(ctx context.Context, tx *gorm.DB, memberID uint32, request *InfiniteScrollRequest) ([]RoomInfo, error) {
	// TODO: 전략 패턴으로 orderBy 및 dir에 따른 repository 메서드 차별화 구현 (time, name, memberCnt 등)

	// Initial request (first page)
	if request.After == DefaultAfter {
		return s.roomRepository.FindRoomInfosByMemberIDInitial(ctx, tx, memberID, InfiniteScrollPageSize)
	}

	// Parse cursor (after) as timestamp
	afterTime, err := time.Parse(time.RFC3339, request.After)
	if err != nil {
		return nil, fmt.Errorf("잘못된 커서 형식: %w", ErrInvalidCursor)
	}

	// Subsequent requests (after cursor)
	return s.roomRepository.FindRoomInfosByMemberIDAfter(ctx, tx, memberID, afterTime, InfiniteScrollPageSize)
}
