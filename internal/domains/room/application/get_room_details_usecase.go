package application

import (
	"context"
	"pray-together/internal/domains/room/domain"
)

// GetRoomDetailsRequest represents the request to get room details
type GetRoomDetailsRequest struct {
	RoomID   uint64
	MemberID uint64 // For access validation
}

// RoomMemberDetail represents member details in a room
type RoomMemberDetail struct {
	MemberID            uint64          `json:"memberId"`
	MemberName          string          `json:"memberName"`
	Role                domain.RoomRole `json:"role"`
	TotalPrayCount      int             `json:"totalPrayCount"`
	ContinuousPrayCount int             `json:"continuousPrayCount"`
	PrayStreak          int             `json:"prayStreak"`
}

// GetRoomDetailsResponse represents the room details response
type GetRoomDetailsResponse struct {
	ID                    uint64             `json:"id"`
	RoomName              string             `json:"roomName"`
	IsPrivate             bool               `json:"isPrivate"`
	IsBlocked             bool               `json:"isBlocked"`
	PrayStartTime         string             `json:"prayStartTime"`
	PrayEndTime           string             `json:"prayEndTime"`
	NotificationStartTime string             `json:"notificationStartTime"`
	NotificationEndTime   string             `json:"notificationEndTime"`
	MemberCount           int                `json:"memberCount"`
	Members               []RoomMemberDetail `json:"members"`
	IsOwner               bool               `json:"isOwner"`
}

// GetRoomDetailsUseCase handles getting room details
type GetRoomDetailsUseCase struct {
	roomService   *domain.Service
	getMemberName func(ctx context.Context, memberID uint64) (string, error)
}

// NewGetRoomDetailsUseCase creates a new GetRoomDetailsUseCase
func NewGetRoomDetailsUseCase(
	roomService *domain.Service,
	getMemberName func(ctx context.Context, memberID uint64) (string, error),
) *GetRoomDetailsUseCase {
	return &GetRoomDetailsUseCase{
		roomService:   roomService,
		getMemberName: getMemberName,
	}
}

// Execute gets room details
func (uc *GetRoomDetailsUseCase) Execute(ctx context.Context, req *GetRoomDetailsRequest) (*GetRoomDetailsResponse, error) {
	// Validate member access
	if err := uc.roomService.ValidateRoomAccess(ctx, req.RoomID, req.MemberID); err != nil {
		return nil, err
	}

	// Get room with members
	room, err := uc.roomService.GetRoomWithMembers(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}

	// Check if requester is owner
	isOwner := false
	for _, member := range room.Members {
		if member.MemberID == req.MemberID && member.IsOwner() {
			isOwner = true
			break
		}
	}

	// Build member details
	memberDetails := make([]RoomMemberDetail, 0, len(room.Members))
	for _, member := range room.Members {
		memberName := ""
		if uc.getMemberName != nil {
			name, _ := uc.getMemberName(ctx, member.MemberID)
			memberName = name
		}

		memberDetails = append(memberDetails, RoomMemberDetail{
			MemberID:            member.MemberID,
			MemberName:          memberName,
			Role:                member.Role,
			TotalPrayCount:      member.TotalPrayCount,
			ContinuousPrayCount: member.ContinuousPrayCount,
			PrayStreak:          member.GetPrayStreak(),
		})
	}

	return &GetRoomDetailsResponse{
		ID:                    room.ID,
		RoomName:              room.RoomName,
		IsPrivate:             room.IsPrivate,
		IsBlocked:             room.IsBlocked,
		PrayStartTime:         room.PrayStartTime,
		PrayEndTime:           room.PrayEndTime,
		NotificationStartTime: room.NotificationStartTime,
		NotificationEndTime:   room.NotificationEndTime,
		MemberCount:           len(room.Members),
		Members:               memberDetails,
		IsOwner:               isOwner,
	}, nil
}
