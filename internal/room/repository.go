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
	MemberCnt      int64     `json:"memberCnt"`      // 멤버 수
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
			mr.created_time as joined_time,
			mr.is_notification,
			0 as member_cnt
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
func (r *RoomRepository) FindRoomInfosByMemberIDAfter(ctx context.Context, db *gorm.DB, memberID int64, after time.Time, limit int) ([]RoomInfo, error) {
	var results []RoomInfo

	err := db.WithContext(ctx).
		Table("member_room mr").
		Select(`
			room.id,
			room.name,
			room.description,
			mr.created_time as joined_time,
			mr.is_notification,
			0 as member_cnt
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

// Create creates a new room
func (r *RoomRepository) Create(ctx context.Context, db *gorm.DB, room *model.Room) error {
	return db.WithContext(ctx).Create(room).Error
}

func (r *RoomRepository) FindByID(ctx context.Context, db *gorm.DB, roomID int64) (*model.Room, error) {
	var room model.Room
	err := db.WithContext(ctx).Model(&model.Room{}).Where("id = ?", roomID).Find(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

// -------------------------------------------------

type MemberCount struct {
	RoomID      int64 // room_id , gorm tag -> mapping table columne
	MemberCount int64 // member_count , you need remove gorm tag
}

type RoomMember struct {
	MemberID    int64
	Name        string
	PhoneNumber string
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

// DeleteByMemberIDAndRoomID deletes a member_room relationship by memberID and roomID
// Returns true if a record was deleted, false if no record was found
func (r *MemberRoomRepository) DeleteByMemberIDAndRoomID(ctx context.Context, db *gorm.DB, memberID int64, roomID int64) (bool, error) {
	result := db.WithContext(ctx).
		Where("member_id = ? AND room_id = ?", memberID, roomID).
		Delete(&model.MemberRoom{})

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}

func (r *MemberRoomRepository) IsExistMemberInRoom(ctx context.Context, db *gorm.DB, memberID int64, roomID int64) (bool, error) {
	var count int64
	err := db.WithContext(ctx).
		Model(&model.MemberRoom{}).
		Where("member_id = ? AND room_id = ?", memberID, roomID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// FindMemberIDsByRoomID fetches all member IDs in a room
func (r *MemberRoomRepository) FindMemberIDsByRoomID(ctx context.Context, db *gorm.DB, roomID int64) ([]int64, error) {
	var memberIDs []int64
	err := db.WithContext(ctx).
		Model(&model.MemberRoom{}).
		Where("room_id = ?", roomID).
		Pluck("member_id", &memberIDs).Error

	if err != nil {
		return nil, err
	}

	return memberIDs, nil
}

func (r *MemberRoomRepository) FindMemberRooms(ctx context.Context, db *gorm.DB, roomID int64) ([]RoomMember, error) {
	var results []RoomMember
	err := db.WithContext(ctx).
		Table("member_room mr").
		Select("m.id as member_id, m.name, m.phone_number").
		Joins("JOIN member m ON mr.member_id = m.id").
		Where("room_id = ?", roomID).
		Order("m.name ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}
