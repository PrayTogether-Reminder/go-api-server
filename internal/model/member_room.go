package model

import (
	"errors"
)

// MemberRoom represents the many-to-many relationship between Member and Room
// This is the join table that also contains additional relationship data
type MemberRoom struct {
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`

	MemberID int64 `gorm:"column:member_id;not null;uniqueIndex:uk_member_room_member_id_room_id"`
	RoomID   int64 `gorm:"column:room_id;not null;uniqueIndex:uk_member_room_member_id_room_id"`

	Member *Member `gorm:"foreignKey:MemberID;constraint:OnDelete:CASCADE"`
	Room   *Room   `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE"`

	Role           RoomRole `gorm:"column:role;type:VARCHAR2(20);not null"`       // 멤버의 방 내 역할
	IsNotification bool     `gorm:"column:is_notification;not null;default:true"` // 알림 수신 여부

	BaseEntity
}

// TableName specifies the table name for MemberRoom
func (*MemberRoom) TableName() string {
	return "member_room"
}

// NewMemberRoom creates a new MemberRoom instance
// Factory method pattern
func NewMemberRoom(memberID, roomID int64, role RoomRole) (*MemberRoom, error) {
	// Validation
	if memberID <= 0 {
		return nil, errors.New("invalid member ID")
	}
	if roomID <= 0 {
		return nil, errors.New("invalid room ID")
	}

	return &MemberRoom{
		MemberID:       memberID,
		RoomID:         roomID,
		Role:           role,
		IsNotification: true, // Default to true
	}, nil
}
