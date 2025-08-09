package domain

import (
	"errors"
	"time"
	"unicode/utf8"
)

// PrayerType represents the type of prayer
type PrayerType string

const (
	PrayerTypePersonal PrayerType = "PERSONAL"
	PrayerTypeShared   PrayerType = "SHARED"
)

// BaseEntity contains common fields for all entities in prayer domain
type BaseEntity struct {
	CreatedAt time.Time  `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`
}

// Prayer represents a prayer entity
type Prayer struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	MemberID   uint64     `gorm:"column:member_id;not null;index" json:"memberId"`
	RoomID     uint64     `gorm:"column:room_id;not null;index" json:"roomId"`
	Content    string     `gorm:"column:content;type:text;not null" json:"content"`
	Type       PrayerType `gorm:"column:type;not null;default:'PERSONAL'" json:"type"`
	IsAnswered bool       `gorm:"column:is_answered;default:false" json:"isAnswered"`
	AnsweredAt *time.Time `gorm:"column:answered_at" json:"answeredAt,omitempty"`
	BaseEntity
}

// TableName specifies the table name for Prayer
func (Prayer) TableName() string {
	return "prayer"
}

// NewPrayer creates a new prayer
func NewPrayer(memberID, roomID uint64, content string, prayerType PrayerType) (*Prayer, error) {
	prayer := &Prayer{
		MemberID: memberID,
		RoomID:   roomID,
		Content:  content,
		Type:     prayerType,
	}

	if err := prayer.Validate(); err != nil {
		return nil, err
	}

	return prayer, nil
}

// Validate validates prayer data
func (p *Prayer) Validate() error {
	if p.MemberID == 0 {
		return ErrInvalidMemberID
	}

	if p.RoomID == 0 {
		return ErrInvalidRoomID
	}

	contentLength := utf8.RuneCountInString(p.Content)
	if contentLength < 1 || contentLength > 1000 {
		return ErrInvalidPrayerContent
	}

	if p.Type != PrayerTypePersonal && p.Type != PrayerTypeShared {
		return ErrInvalidPrayerType
	}

	return nil
}

// UpdateContent updates the prayer content
func (p *Prayer) UpdateContent(content string) error {
	contentLength := utf8.RuneCountInString(content)
	if contentLength < 1 || contentLength > 1000 {
		return ErrInvalidPrayerContent
	}

	p.Content = content
	return nil
}

// MarkAsAnswered marks the prayer as answered
func (p *Prayer) MarkAsAnswered() {
	now := time.Now()
	p.IsAnswered = true
	p.AnsweredAt = &now
}

// MarkAsUnanswered marks the prayer as unanswered
func (p *Prayer) MarkAsUnanswered() {
	p.IsAnswered = false
	p.AnsweredAt = nil
}

// SetType sets the prayer type
func (p *Prayer) SetType(prayerType PrayerType) error {
	if prayerType != PrayerTypePersonal && prayerType != PrayerTypeShared {
		return ErrInvalidPrayerType
	}

	p.Type = prayerType
	return nil
}

// CanBeEditedBy checks if the prayer can be edited by a member
func (p *Prayer) CanBeEditedBy(memberID uint64) bool {
	return p.MemberID == memberID
}

// PrayerInfo represents prayer information
type PrayerInfo struct {
	ID         uint64     `json:"id"`
	MemberID   uint64     `json:"memberId"`
	MemberName string     `json:"memberName,omitempty"`
	RoomID     uint64     `json:"roomId"`
	RoomName   string     `json:"roomName,omitempty"`
	Content    string     `json:"content"`
	Type       PrayerType `json:"type"`
	IsAnswered bool       `json:"isAnswered"`
	AnsweredAt *time.Time `json:"answeredAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// ToInfo converts Prayer to PrayerInfo
func (p *Prayer) ToInfo() *PrayerInfo {
	return &PrayerInfo{
		ID:         p.ID,
		MemberID:   p.MemberID,
		RoomID:     p.RoomID,
		Content:    p.Content,
		Type:       p.Type,
		IsAnswered: p.IsAnswered,
		AnsweredAt: p.AnsweredAt,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

// PrayerSummary represents a prayer summary
type PrayerSummary struct {
	ID         uint64     `json:"id"`
	Content    string     `json:"content"`
	Type       PrayerType `json:"type"`
	IsAnswered bool       `json:"isAnswered"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// ToSummary converts Prayer to PrayerSummary
func (p *Prayer) ToSummary() *PrayerSummary {
	// Truncate content for summary
	content := p.Content
	if utf8.RuneCountInString(content) > 100 {
		runes := []rune(content)
		content = string(runes[:100]) + "..."
	}

	return &PrayerSummary{
		ID:         p.ID,
		Content:    content,
		Type:       p.Type,
		IsAnswered: p.IsAnswered,
		CreatedAt:  p.CreatedAt,
	}
}

// Domain errors
var (
	ErrPrayerNotFound       = errors.New("prayer not found")
	ErrInvalidMemberID      = errors.New("invalid member ID")
	ErrInvalidRoomID        = errors.New("invalid room ID")
	ErrInvalidPrayerContent = errors.New("prayer content must be between 1 and 1000 characters")
	ErrInvalidPrayerType    = errors.New("invalid prayer type")
	ErrUnauthorizedAccess   = errors.New("unauthorized access to prayer")
	ErrAlreadyCompleted     = errors.New("prayer already completed")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrInvalidPrayerTitleID = errors.New("invalid prayer title ID")
)
