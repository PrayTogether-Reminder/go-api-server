package domain

import (
	"context"
	"time"
)

// Service represents room domain service
type Service struct {
	repo Repository
}

// NewService creates a new room service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateRoom creates a new room with the creator as owner
func (s *Service) CreateRoom(
	ctx context.Context,
	creatorID uint64,
	roomName string,
	description string,
	isPrivate bool,
	prayStartTime, prayEndTime string,
	notificationStartTime, notificationEndTime string,
) (*Room, error) {
	// Create room
	room, err := NewRoom(
		roomName,
		description,
		isPrivate,
		prayStartTime, prayEndTime,
		notificationStartTime, notificationEndTime,
	)
	if err != nil {
		return nil, err
	}

	// Save room
	if err := s.repo.CreateRoom(ctx, room); err != nil {
		return nil, err
	}

	// Add creator as owner
	roomMember := NewRoomMember(room.ID, creatorID, RoleOwner)
	if err := s.repo.AddMemberToRoom(ctx, roomMember); err != nil {
		// Rollback room creation if adding owner fails
		_ = s.repo.DeleteRoom(ctx, room.ID)
		return nil, err
	}

	room.Members = []RoomMember{*roomMember}
	return room, nil
}

// GetRoom gets a room by ID
func (s *Service) GetRoom(ctx context.Context, roomID uint64) (*Room, error) {
	room, err := s.repo.FindRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room == nil {
		return nil, ErrRoomNotFound
	}

	if room.IsBlocked {
		return nil, ErrRoomBlocked
	}

	return room, nil
}

// GetRoomWithMembers gets a room with all its members
func (s *Service) GetRoomWithMembers(ctx context.Context, roomID uint64) (*Room, error) {
	room, err := s.repo.FindRoomWithMembers(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room == nil {
		return nil, ErrRoomNotFound
	}

	if room.IsBlocked {
		return nil, ErrRoomBlocked
	}

	return room, nil
}

// GetMemberRooms gets all rooms for a member
func (s *Service) GetMemberRooms(ctx context.Context, memberID uint64) ([]*Room, error) {
	return s.repo.FindRoomsByMemberID(ctx, memberID)
}

// GetMemberRoomsWithCount gets all rooms for a member with member counts
func (s *Service) GetMemberRoomsWithCount(ctx context.Context, memberID uint64) ([]*RoomInfo, error) {
	// Get rooms
	rooms, err := s.repo.FindRoomsByMemberID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	// Convert to RoomInfo and add member counts
	roomInfos := make([]*RoomInfo, len(rooms))
	for i, room := range rooms {
		roomInfos[i] = room.ToInfo()

		// Get member count for each room
		count, err := s.repo.CountRoomMembers(ctx, room.ID)
		if err != nil {
			// Log error but continue
			count = 0
		}
		roomInfos[i].MemberCount = count
	}

	return roomInfos, nil
}

// GetPublicRooms gets public rooms
func (s *Service) GetPublicRooms(ctx context.Context, limit, offset int) ([]*Room, error) {
	return s.repo.FindPublicRooms(ctx, limit, offset)
}

// UpdateRoom updates room information
func (s *Service) UpdateRoom(ctx context.Context, roomID uint64, updaterID uint64, updates map[string]interface{}) (*Room, error) {
	// Get room
	room, err := s.repo.FindRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room == nil {
		return nil, ErrRoomNotFound
	}

	// Check if updater is owner
	roomMember, err := s.repo.FindRoomMember(ctx, roomID, updaterID)
	if err != nil {
		return nil, err
	}

	if roomMember == nil || !roomMember.IsOwner() {
		return nil, ErrNotRoomOwner
	}

	// Apply updates
	if name, ok := updates["roomName"].(string); ok {
		if err := room.UpdateName(name); err != nil {
			return nil, err
		}
	}

	if isPrivate, ok := updates["isPrivate"].(bool); ok {
		room.SetPrivate(isPrivate)
	}

	if prayStart, ok := updates["prayStartTime"].(string); ok {
		if prayEnd, ok := updates["prayEndTime"].(string); ok {
			if err := room.UpdatePrayTime(prayStart, prayEnd); err != nil {
				return nil, err
			}
		}
	}

	if notifStart, ok := updates["notificationStartTime"].(string); ok {
		if notifEnd, ok := updates["notificationEndTime"].(string); ok {
			if err := room.UpdateNotificationTime(notifStart, notifEnd); err != nil {
				return nil, err
			}
		}
	}

	// Save updates
	if err := s.repo.UpdateRoom(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

// DeleteRoom deletes a room
func (s *Service) DeleteRoom(ctx context.Context, roomID uint64, deleterID uint64) error {
	// Check if deleter is owner
	roomMember, err := s.repo.FindRoomMember(ctx, roomID, deleterID)
	if err != nil {
		return err
	}

	if roomMember == nil || !roomMember.IsOwner() {
		return ErrNotRoomOwner
	}

	return s.repo.DeleteRoom(ctx, roomID)
}

// JoinRoom adds a member to a room
func (s *Service) JoinRoom(ctx context.Context, roomID, memberID uint64) error {
	// Check if room exists
	room, err := s.repo.FindRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room == nil {
		return ErrRoomNotFound
	}

	if room.IsBlocked {
		return ErrRoomBlocked
	}

	// Check if member already in room
	exists, err := s.repo.ExistsMemberInRoom(ctx, roomID, memberID)
	if err != nil {
		return err
	}

	if exists {
		return ErrMemberAlreadyInRoom
	}

	// Add member to room
	roomMember := NewRoomMember(roomID, memberID, RoleMember)
	return s.repo.AddMemberToRoom(ctx, roomMember)
}

// LeaveRoom removes a member from a room
func (s *Service) LeaveRoom(ctx context.Context, roomID, memberID uint64) error {
	// Check if member is in room
	roomMember, err := s.repo.FindRoomMember(ctx, roomID, memberID)
	if err != nil {
		return err
	}

	if roomMember == nil {
		return ErrMemberNotInRoom
	}

	// If member is owner, check if there are other members
	if roomMember.IsOwner() {
		count, err := s.repo.CountRoomMembers(ctx, roomID)
		if err != nil {
			return err
		}

		if count > 1 {
			// Transfer ownership to another member
			members, err := s.repo.FindRoomMembers(ctx, roomID)
			if err != nil {
				return err
			}

			for _, m := range members {
				if m.MemberID != memberID {
					m.SetRole(RoleOwner)
					if err := s.repo.UpdateRoomMember(ctx, m); err != nil {
						return err
					}
					break
				}
			}
		}
	}

	return s.repo.RemoveMemberFromRoom(ctx, roomID, memberID)
}

// GetRoomMembers gets all members of a room
func (s *Service) GetRoomMembers(ctx context.Context, roomID uint64) ([]*RoomMember, error) {
	// Check if room exists
	exists, err := s.repo.ExistsRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrRoomNotFound
	}

	return s.repo.FindRoomMembers(ctx, roomID)
}

// UpdateMemberRole updates a member's role in a room
func (s *Service) UpdateMemberRole(ctx context.Context, roomID, updaterID, targetMemberID uint64, newRole RoomRole) error {
	// Check if updater is owner
	updater, err := s.repo.FindRoomMember(ctx, roomID, updaterID)
	if err != nil {
		return err
	}

	if updater == nil || !updater.IsOwner() {
		return ErrNotRoomOwner
	}

	// Get target member
	targetMember, err := s.repo.FindRoomMember(ctx, roomID, targetMemberID)
	if err != nil {
		return err
	}

	if targetMember == nil {
		return ErrMemberNotInRoom
	}

	// Update role
	targetMember.SetRole(newRole)
	return s.repo.UpdateRoomMember(ctx, targetMember)
}

// RecordPray records that a member has prayed in a room
func (s *Service) RecordPray(ctx context.Context, roomID, memberID uint64) error {
	// Get room member
	roomMember, err := s.repo.FindRoomMember(ctx, roomID, memberID)
	if err != nil {
		return err
	}

	if roomMember == nil {
		return ErrMemberNotInRoom
	}

	// Record pray
	roomMember.RecordPray()

	// Update in repository
	return s.repo.UpdateRoomMember(ctx, roomMember)
}

// GetRoomMemberStats gets statistics for a room member
func (s *Service) GetRoomMemberStats(ctx context.Context, roomID, memberID uint64) (*RoomMemberInfo, error) {
	roomMember, err := s.repo.FindRoomMember(ctx, roomID, memberID)
	if err != nil {
		return nil, err
	}

	if roomMember == nil {
		return nil, ErrMemberNotInRoom
	}

	return roomMember.ToInfo(), nil
}

// BlockRoom blocks a room
func (s *Service) BlockRoom(ctx context.Context, roomID uint64) error {
	room, err := s.repo.FindRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room == nil {
		return ErrRoomNotFound
	}

	room.Block()
	return s.repo.UpdateRoom(ctx, room)
}

// UnblockRoom unblocks a room
func (s *Service) UnblockRoom(ctx context.Context, roomID uint64) error {
	room, err := s.repo.FindRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room == nil {
		return ErrRoomNotFound
	}

	room.Unblock()
	return s.repo.UpdateRoom(ctx, room)
}

// ValidateRoomAccess validates if a member has access to a room
func (s *Service) ValidateRoomAccess(ctx context.Context, roomID, memberID uint64) error {
	exists, err := s.repo.ExistsMemberInRoom(ctx, roomID, memberID)
	if err != nil {
		return err
	}

	if !exists {
		return ErrMemberNotInRoom
	}

	return nil
}

// GetMemberRoomsPaginated gets paginated rooms for a member (matching Java fetchRoomsByMember)
func (s *Service) GetMemberRoomsPaginated(ctx context.Context, memberID uint64, after string, dir string, limit int) ([]*RoomInfo, error) {
	// Match Java logic: "0" means first page
	if after == "0" || after == "" {
		// Get first page of rooms ordered by joined time desc
		return s.repo.FindFirstRoomInfosByMemberID(ctx, memberID, limit)
	}

	// Parse time cursor for subsequent pages
	afterTime, err := time.Parse(time.RFC3339Nano, after)
	if err != nil {
		// If parsing fails, return empty to match Java behavior
		return []*RoomInfo{}, nil
	}

	// Get rooms after the specified time
	return s.repo.FindRoomInfosByMemberIDAfterTime(ctx, memberID, afterTime, limit)
}
