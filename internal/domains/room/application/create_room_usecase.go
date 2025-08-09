package application

import (
	"context"
	"pray-together/internal/domains/room/domain"
)

// CreateRoomRequest represents the request to create a room
type CreateRoomRequest struct {
	CreatorID             uint64
	RoomName              string
	Description           string // Added to match Java API
	IsPrivate             bool
	PrayStartTime         string
	PrayEndTime           string
	NotificationStartTime string
	NotificationEndTime   string
}

// CreateRoomResponse represents the response after creating a room
type CreateRoomResponse struct {
	ID                    uint64 `json:"id"`
	RoomName              string `json:"roomName"`
	IsPrivate             bool   `json:"isPrivate"`
	PrayStartTime         string `json:"prayStartTime"`
	PrayEndTime           string `json:"prayEndTime"`
	NotificationStartTime string `json:"notificationStartTime"`
	NotificationEndTime   string `json:"notificationEndTime"`
}

// CreateRoomUseCase handles room creation
type CreateRoomUseCase struct {
	roomService *domain.Service
}

// NewCreateRoomUseCase creates a new CreateRoomUseCase
func NewCreateRoomUseCase(roomService *domain.Service) *CreateRoomUseCase {
	return &CreateRoomUseCase{
		roomService: roomService,
	}
}

// Execute creates a new room
func (uc *CreateRoomUseCase) Execute(ctx context.Context, req *CreateRoomRequest) (*CreateRoomResponse, error) {
	room, err := uc.roomService.CreateRoom(
		ctx,
		req.CreatorID,
		req.RoomName,
		req.Description,
		req.IsPrivate,
		req.PrayStartTime,
		req.PrayEndTime,
		req.NotificationStartTime,
		req.NotificationEndTime,
	)
	if err != nil {
		return nil, err
	}

	return &CreateRoomResponse{
		ID:                    room.ID,
		RoomName:              room.RoomName,
		IsPrivate:             room.IsPrivate,
		PrayStartTime:         room.PrayStartTime,
		PrayEndTime:           room.PrayEndTime,
		NotificationStartTime: room.NotificationStartTime,
		NotificationEndTime:   room.NotificationEndTime,
	}, nil
}
