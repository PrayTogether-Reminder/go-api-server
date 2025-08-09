package domain

import (
	"time"
)

// PrayerCompletion represents a prayer completion record
type PrayerCompletion struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PrayerTitleID uint64     `gorm:"column:prayer_title_id;not null;index" json:"prayerTitleId"`
	MemberID      uint64     `gorm:"column:member_id;not null;index" json:"memberId"`
	CompletedAt   time.Time  `gorm:"column:completed_at;not null" json:"completedAt"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null" json:"updatedAt"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`

	// Relations
	PrayerTitle *PrayerTitle `gorm:"foreignKey:PrayerTitleID" json:"prayerTitle,omitempty"`
}

// TableName specifies the table name for PrayerCompletion
func (PrayerCompletion) TableName() string {
	return "prayer_completion"
}

// NewPrayerCompletion creates a new prayer completion
func NewPrayerCompletion(prayerTitleID, memberID uint64) *PrayerCompletion {
	now := time.Now()
	return &PrayerCompletion{
		PrayerTitleID: prayerTitleID,
		MemberID:      memberID,
		CompletedAt:   now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// IsValid validates the prayer completion
func (p *PrayerCompletion) IsValid() error {
	if p.PrayerTitleID == 0 {
		return ErrInvalidPrayerTitleID
	}
	if p.MemberID == 0 {
		return ErrInvalidMemberID
	}
	return nil
}
