package room

import (
	"context"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

// RoomInfo represents room information with member details
// Created by Repository queries (projection from JOIN)
type RoomInfo struct {
	ID             int64     `json:"id"`             // 방 ID
	Name           string    `json:"name"`           // 방 이름
	Description    string    `json:"description"`    // 방 설명
	JoinedTime     time.Time `json:"joinedTime"`     // 가입 시간
	IsNotification bool      `json:"isNotification"` // 알림 여부
	MemberCount    int64     `json:"memberCount"`    // 멤버 수
}

// ===== RoomRepository =====

type RoomRepository struct{}

func NewRoomRepository() *RoomRepository {
	return &RoomRepository{}
}

// FindRoomInfosByMemberIDInitial fetches the first page of rooms ordered by joined time desc
func (r *RoomRepository) FindRoomInfosByMemberIDInitial(ctx context.Context, db *gorm.DB, memberID int64, limit int) ([]RoomInfo, error) {
	var results []RoomInfo

	err := db.WithContext(ctx).
		Table("member_room mr").
		Select(`
			room.id,
			room.name,
			room.description,
			mr.created_at as joined_time,
			mr.is_notification,
			0 as member_count
		`).
		Joins("JOIN room ON mr.room_id = room.id").
		Where("mr.member_id = ?", memberID).
		Order("mr.created_at DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// FindRoomInfosByMemberIDAfter fetches rooms after a specific cursor (created_time)
func (r *RoomRepository) FindRoomInfosByMemberIDAfter(ctx context.Context, db *gorm.DB, memberID int64, after time.Time, limit int) ([]RoomInfo, error) {
	var results []RoomInfo

	err := db.WithContext(ctx).
		Table("member_room mr").
		Select(`
			room.id,
			room.name,
			room.description,
			mr.created_at as joined_time,
			mr.is_notification,
			0 as member_count
		`).
		Joins("JOIN room ON mr.room_id = room.id").
		Where("mr.member_id = ? AND mr.created_at < ?", memberID, after).
		Order("mr.created_at DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// Create creates a new room
func (r *RoomRepository) Create(ctx context.Context, db *gorm.DB, room *model.Room) error {
	return db.WithContext(ctx).Create(room).Error
}

// -------------------------------------------------

type MemberCount struct {
	RoomID      int64 // room_id , gorm tag -> mapping table columne
	MemberCount int64 // member_count , you need remove gorm tag
}

// MemberRoomRepository handles database operations for member_room relationships

type MemberRoomRepository struct{}

func NewMemberRoomRepository() *MemberRoomRepository {
	return &MemberRoomRepository{}
}

// FindMemberCountsByRoomIDs fetches member counts for given room IDs (batch operation)
func (r *MemberRoomRepository) FindMemberCountsByRoomIDs(ctx context.Context, db *gorm.DB, roomIDs []int64) ([]MemberCount, error) {
	var results []MemberCount

	if len(roomIDs) == 0 {
		return results, nil
	}

	err := db.WithContext(ctx).
		Model(&model.MemberRoom{}).
		Select("room_id, COUNT(*) as member_count").
		Where("room_id IN ?", roomIDs).
		Group("room_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// Create creates a new member_room relationship
func (r *MemberRoomRepository) Create(ctx context.Context, db *gorm.DB, memberRoom *model.MemberRoom) error {
	return db.WithContext(ctx).Create(memberRoom).Error
}
