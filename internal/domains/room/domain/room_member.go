package domain

import (
	"time"
)

// RoomRole represents member role in a room
type RoomRole string

const (
	RoleOwner  RoomRole = "OWNER"
	RoleMember RoomRole = "MEMBER"
)

// RoomMember represents the relationship between Room and Member (member_room entity)
type RoomMember struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RoomID              uint64     `gorm:"column:room_id;not null;index:idx_room_member,unique" json:"roomId"`
	MemberID            uint64     `gorm:"column:member_id;not null;index:idx_room_member,unique" json:"memberId"`
	Role                RoomRole   `gorm:"column:role;not null;default:'MEMBER'" json:"role"`
	IsNotification      bool       `gorm:"column:is_notification;default:true" json:"isNotification"`
	LastPrayTime        *time.Time `gorm:"column:last_pray_time" json:"lastPrayTime,omitempty"`
	TotalPrayCount      int        `gorm:"column:total_pray_count;default:0" json:"totalPrayCount"`
	ContinuousPrayCount int        `gorm:"column:continuous_pray_count;default:0" json:"continuousPrayCount"`
	BaseEntity
}

// TableName specifies the table name for RoomMember
func (RoomMember) TableName() string {
	return "member_room"
}

// NewRoomMember creates a new room member
func NewRoomMember(roomID, memberID uint64, role RoomRole) *RoomMember {
	return &RoomMember{
		RoomID:   roomID,
		MemberID: memberID,
		Role:     role,
	}
}

// IsOwner checks if the member is an owner
func (rm *RoomMember) IsOwner() bool {
	return rm.Role == RoleOwner
}

// SetRole updates the member's role
func (rm *RoomMember) SetRole(role RoomRole) {
	rm.Role = role
}

// RecordPray records that the member has prayed
func (rm *RoomMember) RecordPray() {
	now := time.Now()

	// Check if this is a continuous pray
	if rm.LastPrayTime != nil {
		lastPray := rm.LastPrayTime.Truncate(24 * time.Hour)
		today := now.Truncate(24 * time.Hour)
		yesterday := today.AddDate(0, 0, -1)

		if lastPray.Equal(yesterday) {
			// Prayed yesterday, increment continuous count
			rm.ContinuousPrayCount++
		} else if !lastPray.Equal(today) {
			// Didn't pray yesterday, reset continuous count
			rm.ContinuousPrayCount = 1
		}
		// If lastPray equals today, don't change continuous count
	} else {
		// First pray
		rm.ContinuousPrayCount = 1
	}

	rm.LastPrayTime = &now
	rm.TotalPrayCount++
}

// GetPrayStreak returns the current pray streak
func (rm *RoomMember) GetPrayStreak() int {
	if rm.LastPrayTime == nil {
		return 0
	}

	// Check if the streak is still active
	lastPray := rm.LastPrayTime.Truncate(24 * time.Hour)
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	if lastPray.Equal(today) || lastPray.Equal(yesterday) {
		return rm.ContinuousPrayCount
	}

	// Streak is broken
	return 0
}

// RoomMemberInfo represents room member information
type RoomMemberInfo struct {
	ID                  uint64     `json:"id"`
	RoomID              uint64     `json:"roomId"`
	MemberID            uint64     `json:"memberId"`
	MemberName          string     `json:"memberName,omitempty"`
	Role                RoomRole   `json:"role"`
	LastPrayTime        *time.Time `json:"lastPrayTime,omitempty"`
	TotalPrayCount      int        `json:"totalPrayCount"`
	ContinuousPrayCount int        `json:"continuousPrayCount"`
	PrayStreak          int        `json:"prayStreak"`
}

// ToInfo converts RoomMember to RoomMemberInfo
func (rm *RoomMember) ToInfo() *RoomMemberInfo {
	return &RoomMemberInfo{
		ID:                  rm.ID,
		RoomID:              rm.RoomID,
		MemberID:            rm.MemberID,
		Role:                rm.Role,
		LastPrayTime:        rm.LastPrayTime,
		TotalPrayCount:      rm.TotalPrayCount,
		ContinuousPrayCount: rm.ContinuousPrayCount,
		PrayStreak:          rm.GetPrayStreak(),
	}
}

// RoomMemberStats represents statistics for a room member
type RoomMemberStats struct {
	MemberID            uint64     `json:"memberId"`
	MemberName          string     `json:"memberName"`
	TotalPrayCount      int        `json:"totalPrayCount"`
	ContinuousPrayCount int        `json:"continuousPrayCount"`
	PrayStreak          int        `json:"prayStreak"`
	LastPrayTime        *time.Time `json:"lastPrayTime,omitempty"`
	JoinedAt            time.Time  `json:"joinedAt"`
}

// ToStats converts RoomMember to RoomMemberStats
func (rm *RoomMember) ToStats(memberName string) *RoomMemberStats {
	return &RoomMemberStats{
		MemberID:            rm.MemberID,
		MemberName:          memberName,
		TotalPrayCount:      rm.TotalPrayCount,
		ContinuousPrayCount: rm.ContinuousPrayCount,
		PrayStreak:          rm.GetPrayStreak(),
		LastPrayTime:        rm.LastPrayTime,
		JoinedAt:            rm.CreatedAt,
	}
}
