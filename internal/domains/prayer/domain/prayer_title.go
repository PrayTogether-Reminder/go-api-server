package domain

import (
	"errors"
	"time"
)

// PrayerTitle represents a prayer title/topic entity
type PrayerTitle struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RoomID      uint64     `gorm:"column:room_id;not null;index" json:"roomId"`
	CreatorID   uint64     `gorm:"column:creator_id;not null;index" json:"creatorId"`
	Title       string     `gorm:"column:title;not null" json:"title"`
	Description string     `gorm:"column:description;type:text" json:"description,omitempty"`
	IsAnswered  bool       `gorm:"column:is_answered;default:false" json:"isAnswered"`
	AnsweredAt  *time.Time `gorm:"column:answered_at" json:"answeredAt,omitempty"`
	BaseEntity

	// Relations
	Contents []PrayerContent `gorm:"foreignKey:PrayerTitleID" json:"contents,omitempty"`
}

// TableName specifies the table name for PrayerTitle
func (PrayerTitle) TableName() string {
	return "prayer_title"
}

// NewPrayerTitle creates a new prayer title
func NewPrayerTitle(roomID, creatorID uint64, title string) *PrayerTitle {
	return &PrayerTitle{
		RoomID:      roomID,
		CreatorID:   creatorID,
		Title:       title,
		Description: "",
		IsAnswered:  false,
	}
}

// Validate validates prayer title data
func (p *PrayerTitle) Validate() error {
	if p.RoomID == 0 {
		return ErrInvalidRoomID
	}

	if p.CreatorID == 0 {
		return ErrInvalidMemberID
	}

	if p.Title == "" || len(p.Title) > 200 {
		return ErrInvalidTitle
	}

	if len(p.Description) > 1000 {
		return ErrDescriptionTooLong
	}

	return nil
}

// MarkAsAnswered marks the prayer title as answered
func (p *PrayerTitle) MarkAsAnswered() {
	p.IsAnswered = true
	now := time.Now()
	p.AnsweredAt = &now
}

// CanEdit checks if a member can edit this prayer title
func (p *PrayerTitle) CanEdit(memberID uint64) bool {
	return p.CreatorID == memberID
}

// PrayerTitleInfo represents prayer title information for responses (matching Java PrayerTitleInfo)
type PrayerTitleInfo struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	CreatedTime time.Time `json:"createdTime"`
}

// ToPrayerTitleInfo converts PrayerTitle to PrayerTitleInfo (matching Java)
func (p *PrayerTitle) ToPrayerTitleInfo() *PrayerTitleInfo {
	return &PrayerTitleInfo{
		ID:          p.ID,
		Title:       p.Title,
		CreatedTime: p.CreatedAt,
	}
}

// PrayerTitleDetailInfo represents detailed prayer title information (legacy)
type PrayerTitleDetailInfo struct {
	ID           uint64              `json:"id"`
	RoomID       uint64              `json:"roomId"`
	RoomName     string              `json:"roomName,omitempty"`
	CreatorID    uint64              `json:"creatorId"`
	CreatorName  string              `json:"creatorName,omitempty"`
	Title        string              `json:"title"`
	Description  string              `json:"description,omitempty"`
	IsAnswered   bool                `json:"isAnswered"`
	AnsweredAt   *time.Time          `json:"answeredAt,omitempty"`
	ContentCount int                 `json:"contentCount"`
	Contents     []PrayerContentInfo `json:"contents,omitempty"`
	CreatedAt    time.Time           `json:"createdAt"`
	UpdatedAt    time.Time           `json:"updatedAt"`
}

// ToDetailInfo converts PrayerTitle to PrayerTitleDetailInfo (legacy)
func (p *PrayerTitle) ToDetailInfo() *PrayerTitleDetailInfo {
	info := &PrayerTitleDetailInfo{
		ID:           p.ID,
		RoomID:       p.RoomID,
		CreatorID:    p.CreatorID,
		Title:        p.Title,
		Description:  p.Description,
		IsAnswered:   p.IsAnswered,
		AnsweredAt:   p.AnsweredAt,
		ContentCount: len(p.Contents),
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}

	// Convert contents if loaded
	if len(p.Contents) > 0 {
		info.Contents = make([]PrayerContentInfo, len(p.Contents))
		for i, content := range p.Contents {
			info.Contents[i] = *content.ToInfo()
		}
	}

	return info
}

// Additional prayer title errors
var (
	ErrInvalidTitle        = errors.New("invalid prayer title")
	ErrDescriptionTooLong  = errors.New("description too long")
	ErrPrayerTitleNotFound = errors.New("prayer title not found")
)
