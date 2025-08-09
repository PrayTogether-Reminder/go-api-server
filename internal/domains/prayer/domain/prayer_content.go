package domain

import (
	"errors"
	"time"
)

// PrayerContent represents individual prayer content within a prayer title
type PrayerContent struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PrayerTitleID uint64     `gorm:"column:prayer_title_id;not null;index" json:"prayerTitleId"`
	AuthorID      uint64     `gorm:"column:author_id;not null;index" json:"authorId"`
	Content       string     `gorm:"column:content;type:text;not null" json:"content"`
	IsCompleted   bool       `gorm:"column:is_completed;default:false" json:"isCompleted"`
	CompletedAt   *time.Time `gorm:"column:completed_at" json:"completedAt,omitempty"`
	BaseEntity
}

// TableName specifies the table name for PrayerContent
func (PrayerContent) TableName() string {
	return "prayer_content"
}

// NewPrayerContent creates a new prayer content
func NewPrayerContent(prayerTitleID, authorID uint64, content string) *PrayerContent {
	return &PrayerContent{
		PrayerTitleID: prayerTitleID,
		AuthorID:      authorID,
		Content:       content,
		IsCompleted:   false,
	}
}

// Validate validates prayer content data
func (p *PrayerContent) Validate() error {
	if p.PrayerTitleID == 0 {
		return ErrInvalidPrayerTitleID
	}

	if p.AuthorID == 0 {
		return ErrInvalidMemberID
	}

	if p.Content == "" || len(p.Content) > 5000 {
		return ErrInvalidContent
	}

	return nil
}

// MarkAsCompleted marks the prayer content as completed
func (p *PrayerContent) MarkAsCompleted() {
	p.IsCompleted = true
	now := time.Now()
	p.CompletedAt = &now
}

// CanEdit checks if a member can edit this prayer content
func (p *PrayerContent) CanEdit(memberID uint64) bool {
	return p.AuthorID == memberID
}

// PrayerContentInfo represents prayer content information
type PrayerContentInfo struct {
	ID            uint64     `json:"id"`
	PrayerTitleID uint64     `json:"prayerTitleId"`
	AuthorID      uint64     `json:"authorId"`
	AuthorName    string     `json:"authorName,omitempty"`
	Content       string     `json:"content"`
	IsCompleted   bool       `json:"isCompleted"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// ToInfo converts PrayerContent to PrayerContentInfo
func (p *PrayerContent) ToInfo() *PrayerContentInfo {
	return &PrayerContentInfo{
		ID:            p.ID,
		PrayerTitleID: p.PrayerTitleID,
		AuthorID:      p.AuthorID,
		Content:       p.Content,
		IsCompleted:   p.IsCompleted,
		CompletedAt:   p.CompletedAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// Additional prayer content errors
var (
	ErrInvalidContent        = errors.New("invalid prayer content")
	ErrPrayerContentNotFound = errors.New("prayer content not found")
)
