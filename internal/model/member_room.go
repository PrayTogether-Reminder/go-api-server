package model

// MemberRoom represents the relationship between a member and a room
// This is a junction table (Java의 MemberRoom 엔티티와 동일)
// Oracle IDENTITY (auto-increment) is used for ID generation
type MemberRoom struct {
	// Primary key - Oracle IDENTITY (auto-increment)
	ID uint32 `gorm:"column:id;primaryKey;autoIncrement"`

	// Foreign keys
	MemberID uint32 `gorm:"column:member_id;not null;uniqueIndex:idx_member_room_member_id_room_id"` // FK to member
	RoomID   uint32 `gorm:"column:room_id;not null;uniqueIndex:idx_member_room_member_id_room_id"`   // FK to room

	// Foreign key relations (Lazy loading in Java)
	Member *Member `gorm:"foreignKey:MemberID;references:ID;constraint:OnDelete:CASCADE"` // Member 엔티티 참조
	Room   *Room   `gorm:"foreignKey:RoomID;references:ID;constraint:OnDelete:CASCADE"`   // Room 엔티티 참조

	// Additional fields
	Role           string `gorm:"column:role;type:VARCHAR2(10);not null"`       // 역할 (ADMIN, MEMBER 등)
	IsNotification bool   `gorm:"column:is_notification;not null;default:true"` // 알림 여부

	BaseEntity
}

// TableName specifies the table name for MemberRoom
func (*MemberRoom) TableName() string {
	return "member_room"
}

// NewRoomMember creates a new MemberRoom instance
func NewRoomMember(memberID, roomID uint32, role string, isNotification bool) *MemberRoom {
	return &MemberRoom{
		MemberID:       memberID,
		RoomID:         roomID,
		Role:           role,
		IsNotification: isNotification,
	}
}

const (
	RoomRoleAdmin  = "ADMIN"  // 방 관리자
	RoomRoleMember = "MEMBER" // 일반 멤버
)
