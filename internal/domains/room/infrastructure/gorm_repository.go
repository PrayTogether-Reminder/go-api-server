package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"pray-together/internal/domains/room/domain"
)

// GormRepository implements room domain repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM repository
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{
		db: db,
	}
}

// Room operations

// CreateRoom creates a new room
func (r *GormRepository) CreateRoom(ctx context.Context, room *domain.Room) error {
	return r.db.WithContext(ctx).Create(room).Error
}

// FindRoomByID finds a room by ID
func (r *GormRepository) FindRoomByID(ctx context.Context, id uint64) (*domain.Room, error) {
	var room domain.Room
	err := r.db.WithContext(ctx).Where("id = ? AND is_blocked = ?", id, false).First(&room).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &room, err
}

// FindRoomsByMemberID finds rooms by member ID
func (r *GormRepository) FindRoomsByMemberID(ctx context.Context, memberID uint64) ([]*domain.Room, error) {
	var rooms []*domain.Room

	err := r.db.WithContext(ctx).
		Table("room").
		Joins("JOIN member_room ON room.id = member_room.room_id").
		Where("member_room.member_id = ? AND room.is_blocked = ?", memberID, false).
		Find(&rooms).Error

	return rooms, err
}

// FindPublicRooms finds public rooms
func (r *GormRepository) FindPublicRooms(ctx context.Context, limit, offset int) ([]*domain.Room, error) {
	var rooms []*domain.Room

	err := r.db.WithContext(ctx).
		Where("is_private = ? AND is_blocked = ?", false, false).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&rooms).Error

	return rooms, err
}

// UpdateRoom updates a room
func (r *GormRepository) UpdateRoom(ctx context.Context, room *domain.Room) error {
	return r.db.WithContext(ctx).Save(room).Error
}

// DeleteRoom deletes a room
func (r *GormRepository) DeleteRoom(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete all room members first
		if err := tx.Where("room_id = ?", id).Delete(&domain.RoomMember{}).Error; err != nil {
			return err
		}

		// Delete the room
		return tx.Delete(&domain.Room{}, id).Error
	})
}

// ExistsRoomByID checks if room exists by ID
func (r *GormRepository) ExistsRoomByID(ctx context.Context, id uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Room{}).
		Where("id = ? AND is_blocked = ?", id, false).
		Count(&count).Error

	return count > 0, err
}

// RoomMember operations

// AddMemberToRoom adds a member to a room
func (r *GormRepository) AddMemberToRoom(ctx context.Context, roomMember *domain.RoomMember) error {
	return r.db.WithContext(ctx).Create(roomMember).Error
}

// RemoveMemberFromRoom removes a member from a room
func (r *GormRepository) RemoveMemberFromRoom(ctx context.Context, roomID, memberID uint64) error {
	return r.db.WithContext(ctx).
		Where("room_id = ? AND member_id = ?", roomID, memberID).
		Delete(&domain.RoomMember{}).Error
}

// FindRoomMember finds a room member
func (r *GormRepository) FindRoomMember(ctx context.Context, roomID, memberID uint64) (*domain.RoomMember, error) {
	var roomMember domain.RoomMember
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND member_id = ?", roomID, memberID).
		First(&roomMember).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &roomMember, err
}

// FindRoomMembers finds all members of a room
func (r *GormRepository) FindRoomMembers(ctx context.Context, roomID uint64) ([]*domain.RoomMember, error) {
	var members []*domain.RoomMember

	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Find(&members).Error

	return members, err
}

// UpdateRoomMember updates a room member
func (r *GormRepository) UpdateRoomMember(ctx context.Context, roomMember *domain.RoomMember) error {
	return r.db.WithContext(ctx).Save(roomMember).Error
}

// CountRoomMembers counts members in a room
func (r *GormRepository) CountRoomMembers(ctx context.Context, roomID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.RoomMember{}).
		Where("room_id = ?", roomID).
		Count(&count).Error

	return int(count), err
}

// ExistsMemberInRoom checks if member exists in room
func (r *GormRepository) ExistsMemberInRoom(ctx context.Context, roomID, memberID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.RoomMember{}).
		Where("room_id = ? AND member_id = ?", roomID, memberID).
		Count(&count).Error

	return count > 0, err
}

// Advanced queries

// FindRoomWithMembers finds a room with all its members
func (r *GormRepository) FindRoomWithMembers(ctx context.Context, roomID uint64) (*domain.Room, error) {
	var room domain.Room
	err := r.db.WithContext(ctx).
		Preload("Members").
		Where("id = ? AND is_blocked = ?", roomID, false).
		First(&room).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &room, err
}

// FindRoomsByName finds rooms by name pattern
func (r *GormRepository) FindRoomsByName(ctx context.Context, name string) ([]*domain.Room, error) {
	var rooms []*domain.Room

	err := r.db.WithContext(ctx).
		Where("room_name LIKE ? AND is_blocked = ?", "%"+name+"%", false).
		Find(&rooms).Error

	return rooms, err
}

// FindActiveRooms finds active rooms (rooms with recent activity)
func (r *GormRepository) FindActiveRooms(ctx context.Context, limit, offset int) ([]*domain.Room, error) {
	var rooms []*domain.Room

	// This would typically join with prayer or activity tables
	err := r.db.WithContext(ctx).
		Where("is_blocked = ?", false).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rooms).Error

	return rooms, err
}

// FindRoomMemberIDs finds all member IDs in a room
func (r *GormRepository) FindRoomMemberIDs(ctx context.Context, roomID uint64) ([]uint64, error) {
	var memberIDs []uint64
	err := r.db.WithContext(ctx).
		Model(&domain.RoomMember{}).
		Where("room_id = ? AND deleted_at IS NULL", roomID).
		Pluck("member_id", &memberIDs).Error

	return memberIDs, err
}

// GetRoomOwner gets the owner of a room
func (r *GormRepository) GetRoomOwner(ctx context.Context, roomID uint64) (*domain.RoomMember, error) {
	var owner domain.RoomMember
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND role = ?", roomID, domain.RoleOwner).
		First(&owner).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &owner, err
}

// FindFirstRoomInfosByMemberID finds first page of rooms for a member (matching Java)
func (r *GormRepository) FindFirstRoomInfosByMemberID(ctx context.Context, memberID uint64, limit int) ([]*domain.RoomInfo, error) {
	var roomInfos []*domain.RoomInfo

	// Match Java: SELECT r.*, mr.created_at as joined_time FROM room r
	// JOIN member_room mr ON r.id = mr.room_id
	// WHERE mr.member_id = ? AND r.is_blocked = false
	// ORDER BY mr.created_at DESC
	// LIMIT ?
	err := r.db.WithContext(ctx).
		Table("room r").
		Select(`r.id, r.room_name as name, r.description, r.created_at, 
		        mr.created_at as joined_time, mr.role, mr.is_notification,
		        (SELECT COUNT(*) FROM member_room WHERE room_id = r.id) as member_count`).
		Joins("JOIN member_room mr ON r.id = mr.room_id").
		Where("mr.member_id = ? AND r.is_blocked = ?", memberID, false).
		Order("mr.created_at DESC").
		Limit(limit).
		Scan(&roomInfos).Error

	return roomInfos, err
}

// FindRoomInfosByMemberIDAfterTime finds rooms for a member after a specific time (matching Java)
func (r *GormRepository) FindRoomInfosByMemberIDAfterTime(ctx context.Context, memberID uint64, afterTime time.Time, limit int) ([]*domain.RoomInfo, error) {
	var roomInfos []*domain.RoomInfo

	// Match Java: SELECT r.*, mr.created_at as joined_time FROM room r
	// JOIN member_room mr ON r.id = mr.room_id
	// WHERE mr.member_id = ? AND r.is_blocked = false AND mr.created_at < ?
	// ORDER BY mr.created_at DESC
	// LIMIT ?
	err := r.db.WithContext(ctx).
		Table("room r").
		Select(`r.id, r.room_name as name, r.description, r.created_at,
		        mr.created_at as joined_time, mr.role, mr.is_notification,
		        (SELECT COUNT(*) FROM member_room WHERE room_id = r.id) as member_count`).
		Joins("JOIN member_room mr ON r.id = mr.room_id").
		Where("mr.member_id = ? AND r.is_blocked = ? AND mr.created_at < ?", memberID, false, afterTime).
		Order("mr.created_at DESC").
		Limit(limit).
		Scan(&roomInfos).Error

	return roomInfos, err
}
