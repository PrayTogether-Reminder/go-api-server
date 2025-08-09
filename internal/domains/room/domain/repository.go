package domain

import (
	"context"
	"time"
)

// Repository interface for room domain
type Repository interface {
	// Room operations
	CreateRoom(ctx context.Context, room *Room) error
	FindRoomByID(ctx context.Context, id uint64) (*Room, error)
	FindRoomsByMemberID(ctx context.Context, memberID uint64) ([]*Room, error)
	FindPublicRooms(ctx context.Context, limit, offset int) ([]*Room, error)
	UpdateRoom(ctx context.Context, room *Room) error
	DeleteRoom(ctx context.Context, id uint64) error
	ExistsRoomByID(ctx context.Context, id uint64) (bool, error)

	// RoomMember operations
	AddMemberToRoom(ctx context.Context, roomMember *RoomMember) error
	RemoveMemberFromRoom(ctx context.Context, roomID, memberID uint64) error
	FindRoomMember(ctx context.Context, roomID, memberID uint64) (*RoomMember, error)
	FindRoomMembers(ctx context.Context, roomID uint64) ([]*RoomMember, error)
	FindRoomMemberIDs(ctx context.Context, roomID uint64) ([]uint64, error)
	UpdateRoomMember(ctx context.Context, roomMember *RoomMember) error
	CountRoomMembers(ctx context.Context, roomID uint64) (int, error)
	ExistsMemberInRoom(ctx context.Context, roomID, memberID uint64) (bool, error)

	// Advanced queries
	FindRoomWithMembers(ctx context.Context, roomID uint64) (*Room, error)
	FindRoomsByName(ctx context.Context, name string) ([]*Room, error)
	FindActiveRooms(ctx context.Context, limit, offset int) ([]*Room, error)
	GetRoomOwner(ctx context.Context, roomID uint64) (*RoomMember, error)

	// Pagination queries (matching Java)
	FindFirstRoomInfosByMemberID(ctx context.Context, memberID uint64, limit int) ([]*RoomInfo, error)
	FindRoomInfosByMemberIDAfterTime(ctx context.Context, memberID uint64, afterTime time.Time, limit int) ([]*RoomInfo, error)
}
