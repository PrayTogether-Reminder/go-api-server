package application

import (
	"context"
	"pray-together/internal/domains/room/domain"
)

// JoinRoomRequest represents the request to join a room
type JoinRoomRequest struct {
	RoomID   uint64
	MemberID uint64
}

// JoinRoomResponse represents the response after joining a room
type JoinRoomResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// JoinRoomUseCase handles room joining
type JoinRoomUseCase struct {
	roomService *domain.Service
}

// NewJoinRoomUseCase creates a new JoinRoomUseCase
func NewJoinRoomUseCase(roomService *domain.Service) *JoinRoomUseCase {
	return &JoinRoomUseCase{
		roomService: roomService,
	}
}

// Execute adds a member to a room
func (uc *JoinRoomUseCase) Execute(ctx context.Context, req *JoinRoomRequest) (*JoinRoomResponse, error) {
	err := uc.roomService.JoinRoom(ctx, req.RoomID, req.MemberID)
	if err != nil {
		return nil, err
	}

	return &JoinRoomResponse{
		Success: true,
		Message: "Successfully joined the room",
	}, nil
}
