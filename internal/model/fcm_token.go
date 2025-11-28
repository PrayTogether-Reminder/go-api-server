package model

// FcmToken represents a Firebase token registered by a member.
type FcmToken struct {
	ID       int64   `gorm:"column:id;primaryKey;autoIncrement"`
	MemberID int64   `gorm:"column:member_id;not null"`
	Member   *Member `gorm:"foreignKey:MemberID;constraint:OnDelete:CASCADE"`
	Token    string  `gorm:"column:token;type:VARCHAR2(512);not null"`
	IsActive bool    `gorm:"column:is_active;not null"`

	BaseEntity
}

// TableName returns the table name.
func (*FcmToken) TableName() string {
	return "fcm_token"
}
