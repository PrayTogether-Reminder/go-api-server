package interfaces

import (
	"context"
	"pray-together/internal/domains/room/domain"
)

// API represents the public interface for room domain
// Other domains should use this interface instead of accessing internal components
type API interface {
	// Room operations
	GetRoomInfo(ctx context.Context, roomID uint64) (*domain.RoomInfo, error)
	ValidateRoomAccess(ctx context.Context, roomID, memberID uint64) error
	GetMemberRooms(ctx context.Context, memberID uint64) ([]*domain.RoomInfo, error)

	// Member operations
	IsMemberInRoom(ctx context.Context, roomID, memberID uint64) (bool, error)
	IsRoomOwner(ctx context.Context, roomID, memberID uint64) (bool, error)
	GetRoomMemberInfo(ctx context.Context, roomID, memberID uint64) (*domain.RoomMemberInfo, error)

	// Pray operations
	RecordMemberPray(ctx context.Context, roomID, memberID uint64) error
}

// Module represents the room module with all its components
type Module struct {
	Service    *domain.Service
	Repository domain.Repository
}

// NewModule creates a new room module
func NewModule(repo domain.Repository) *Module {
	return &Module{
		Service:    domain.NewService(repo),
		Repository: repo,
	}
}

// GetRoomInfo implements API interface
func (m *Module) GetRoomInfo(ctx context.Context, roomID uint64) (*domain.RoomInfo, error) {
	room, err := m.Service.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	return room.ToInfo(), nil
}

// ValidateRoomAccess implements API interface
func (m *Module) ValidateRoomAccess(ctx context.Context, roomID, memberID uint64) error {
	return m.Service.ValidateRoomAccess(ctx, roomID, memberID)
}

// GetMemberRooms implements API interface
func (m *Module) GetMemberRooms(ctx context.Context, memberID uint64) ([]*domain.RoomInfo, error) {
	rooms, err := m.Service.GetMemberRooms(ctx, memberID)
	if err != nil {
		return nil, err
	}

	roomInfos := make([]*domain.RoomInfo, len(rooms))
	for i, room := range rooms {
		roomInfos[i] = room.ToInfo()
	}

	return roomInfos, nil
}

// IsMemberInRoom implements API interface
func (m *Module) IsMemberInRoom(ctx context.Context, roomID, memberID uint64) (bool, error) {
	err := m.Service.ValidateRoomAccess(ctx, roomID, memberID)
	if err != nil {
		if err == domain.ErrMemberNotInRoom {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsRoomOwner implements API interface
func (m *Module) IsRoomOwner(ctx context.Context, roomID, memberID uint64) (bool, error) {
	stats, err := m.Service.GetRoomMemberStats(ctx, roomID, memberID)
	if err != nil {
		if err == domain.ErrMemberNotInRoom {
			return false, nil
		}
		return false, err
	}
	return stats.Role == domain.RoleOwner, nil
}

// GetRoomMemberInfo implements API interface
func (m *Module) GetRoomMemberInfo(ctx context.Context, roomID, memberID uint64) (*domain.RoomMemberInfo, error) {
	return m.Service.GetRoomMemberStats(ctx, roomID, memberID)
}

// RecordMemberPray implements API interface
func (m *Module) RecordMemberPray(ctx context.Context, roomID, memberID uint64) error {
	return m.Service.RecordPray(ctx, roomID, memberID)
}

// JoinRoom allows a member to join a room
func (m *Module) JoinRoom(ctx context.Context, roomID, memberID uint64) error {
	return m.Service.JoinRoom(ctx, roomID, memberID)
}

// GetRoomDetails returns detailed room information
func (m *Module) GetRoomDetails(ctx context.Context, roomID uint64) (*domain.Room, error) {
	return m.Service.GetRoom(ctx, roomID)
}

// GetRoomMemberIDs returns all member IDs in a room
func (m *Module) GetRoomMemberIDs(ctx context.Context, roomID uint64) ([]uint64, error) {
	return m.Repository.FindRoomMemberIDs(ctx, roomID)
}
