package room

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// RoomInfo represents room information with member details
// Created by Repository queries (projection from JOIN)
type RoomInfo struct {
	ID             uint32    `json:"id"`             // 방 ID
	Name           string    `json:"name"`           // 방 이름
	Description    string    `json:"description"`    // 방 설명
	JoinedTime     time.Time `json:"joinedTime"`     // 가입 시간
	IsNotification bool      `json:"isNotification"` // 알림 여부
	MemberCount    int       `json:"memberCount"`    // 멤버 수
}

// ===== RoomRepository =====

type RoomRepository struct{}

func NewRoomRepository() *RoomRepository {
	return &RoomRepository{}
}

// FindRoomInfosByMemberIDInitial fetches the first page of rooms ordered by joined time desc
func (r *RoomRepository) FindRoomInfosByMemberIDInitial(ctx context.Context, db *gorm.DB, memberID uint32, limit int) ([]RoomInfo, error) {
	var results []RoomInfo

	err := db.WithContext(ctx).
		Table("member_room mr").
		Select(`
			room.id,
			room.name,
			room.description,
			mr.created_time as joined_time,
			mr.is_notification,
			0 as member_count
		`).
		Joins("JOIN room ON mr.room_id = room.id").
		Where("mr.member_id = ?", memberID).
		Order("mr.created_time DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// FindRoomInfosByMemberIDAfter fetches rooms after a specific cursor (created_time)
func (r *RoomRepository) FindRoomInfosByMemberIDAfter(ctx context.Context, db *gorm.DB, memberID uint32, after time.Time, limit int) ([]RoomInfo, error) {
	var results []RoomInfo

	err := db.WithContext(ctx).
		Table("member_room mr").
		Select(`
			room.id,
			room.name,
			room.description,
			mr.created_time as joined_time,
			mr.is_notification,
			0 as member_count
		`).
		Joins("JOIN room ON mr.room_id = room.id").
		Where("mr.member_id = ? AND mr.created_time < ?", memberID, after).
		Order("mr.created_time DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// -------------------------------------------------

type MemberCount struct {
	RoomID      uint32 `gorm:"column:room_id"`
	MemberCount int    `gorm:"column:member_count"`
}

// MemberRoomRepository handles database operations for member_room relationships

type MemberRoomRepository struct{}

func NewMemberRoomRepository() *MemberRoomRepository {
	return &MemberRoomRepository{}
}

// FindMemberCountsByRoomIDs fetches member counts for given room IDs (batch operation)
func (r *MemberRoomRepository) FindMemberCountsByRoomIDs(ctx context.Context, db *gorm.DB, roomIDs []uint32) ([]MemberCount, error) {
	var results []MemberCount

	if len(roomIDs) == 0 {
		return results, nil
	}

	err := db.WithContext(ctx).
		Table("member_room").
		Select("room_id, COUNT(*) as member_count").
		Where("room_id IN ?", roomIDs).
		Group("room_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}
