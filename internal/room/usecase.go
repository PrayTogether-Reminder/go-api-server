package room

import (
	"context"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"gorm.io/gorm"
)

// RoomUseCase coordinates room-related application logic with transaction management.
type RoomUseCase struct {
	db          *gorm.DB
	roomService *RoomService
}

func NewRoomUseCase(db *gorm.DB, roomService *RoomService) *RoomUseCase {
	return &RoomUseCase{
		db:          db,
		roomService: roomService,
	}
}

func (u *RoomUseCase) FetchInfiniteScroll(ctx context.Context, memberID int64, request *InfiniteScrollRequest) (*InfiniteScrollResponse, error) {
	var response *InfiniteScrollResponse
	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		result, err := u.roomService.FetchInfiniteScroll(ctx, tx, memberID, request)
		if err != nil {
			return err
		}
		response = result
		return nil
	})
	return response, err
}

func (u *RoomUseCase) CreateRoom(ctx context.Context, memberID int64, request *CreateRoomRequest) (*sharedHttp.MessageResponse, error) {
	var response *sharedHttp.MessageResponse
	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		result, err := u.roomService.CreateRoom(ctx, tx, memberID, request)
		if err != nil {
			return err
		}
		response = result
		return nil
	})
	return response, err
}

func (u *RoomUseCase) ExitRoom(ctx context.Context, memberID int64, roomID int64) (*sharedHttp.MessageResponse, error) {
	var response *sharedHttp.MessageResponse
	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		result, err := u.roomService.ExitRoom(ctx, tx, memberID, roomID)
		if err != nil {
			return err
		}
		response = result
		return nil
	})
	return response, err
}

func (u *RoomUseCase) FetchMembersInRoom(ctx context.Context, memberID int64, roomID int64) (*FetchRoomMemberResponse, error) {
	var response *FetchRoomMemberResponse
	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		result, err := u.roomService.FetchMembersInRoom(ctx, tx, memberID, roomID)
		if err != nil {
			return err
		}
		response = result
		return nil
	})
	return response, err
}
