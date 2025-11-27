package model

import "time"

// RefreshToken represents a refresh token stored in the database
// Each member can have only one refresh token at a time
type RefreshToken struct {
	BaseEntity
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	MemberID  int64     `gorm:"column:member_id;not null;uniqueIndex:idx_refresh_token_member_id"`
	Token     string    `gorm:"column:token;type:VARCHAR2(512);not null"`
	ExpiredAt time.Time `gorm:"column:expired_time;not null"`

	Member *Member `gorm:"foreignKey:MemberID;references:ID;constraint:OnDelete:CASCADE"` // Member 엔티티 참조
}

// TableName specifies the table name for RefreshToken
func (*RefreshToken) TableName() string {
	return "refresh_token"
}

// IsExpired checks if the refresh token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiredAt)
}

// NewRefreshToken creates a new RefreshToken instance
func NewRefreshToken(memberID int64, token string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		MemberID:  memberID,
		Token:     token,
		ExpiredAt: expiresAt,
	}
}
