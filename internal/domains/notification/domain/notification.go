package domain

import (
	"errors"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypePrayer     NotificationType = "PRAYER"
	NotificationTypeInvitation NotificationType = "INVITATION"
	NotificationTypeRoom       NotificationType = "ROOM"
	NotificationTypeSystem     NotificationType = "SYSTEM"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	StatusPending   NotificationStatus = "PENDING"
	StatusSent      NotificationStatus = "SENT"
	StatusFailed    NotificationStatus = "FAILED"
	StatusDelivered NotificationStatus = "DELIVERED"
	StatusRead      NotificationStatus = "READ"
)

// BaseEntity contains common fields for all entities in notification domain
type BaseEntity struct {
	CreatedAt time.Time  `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`
}

// Notification represents a notification entity
type Notification struct {
	ID            uint64             `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	MemberID      uint64             `gorm:"column:member_id;not null;index" json:"memberId"`
	Type          NotificationType   `gorm:"column:type;not null" json:"type"`
	Title         string             `gorm:"column:title;not null" json:"title"`
	Body          string             `gorm:"column:body;not null" json:"body"`
	Data          string             `gorm:"column:data;type:text" json:"data,omitempty"` // JSON string
	Status        NotificationStatus `gorm:"column:status;not null;default:'PENDING'" json:"status"`
	SentAt        *time.Time         `gorm:"column:sent_at" json:"sentAt,omitempty"`
	DeliveredAt   *time.Time         `gorm:"column:delivered_at" json:"deliveredAt,omitempty"`
	ReadAt        *time.Time         `gorm:"column:read_at" json:"readAt,omitempty"`
	FailureReason string             `gorm:"column:failure_reason" json:"failureReason,omitempty"`
	BaseEntity
}

// TableName specifies the table name for Notification
func (Notification) TableName() string {
	return "notification"
}

// NewNotification creates a new notification
func NewNotification(memberID uint64, notifType NotificationType, title, body, data string) (*Notification, error) {
	notification := &Notification{
		MemberID: memberID,
		Type:     notifType,
		Title:    title,
		Body:     body,
		Data:     data,
		Status:   StatusPending,
	}

	if err := notification.Validate(); err != nil {
		return nil, err
	}

	return notification, nil
}

// Validate validates notification data
func (n *Notification) Validate() error {
	if n.MemberID == 0 {
		return ErrInvalidMemberID
	}

	if n.Title == "" || len(n.Title) > 100 {
		return ErrInvalidTitle
	}

	if n.Body == "" || len(n.Body) > 500 {
		return ErrInvalidBody
	}

	if !n.isValidType() {
		return ErrInvalidNotificationType
	}

	return nil
}

// isValidType checks if the notification type is valid
func (n *Notification) isValidType() bool {
	switch n.Type {
	case NotificationTypePrayer, NotificationTypeInvitation, NotificationTypeRoom, NotificationTypeSystem:
		return true
	default:
		return false
	}
}

// MarkAsSent marks the notification as sent
func (n *Notification) MarkAsSent() {
	now := time.Now()
	n.Status = StatusSent
	n.SentAt = &now
}

// MarkAsDelivered marks the notification as delivered
func (n *Notification) MarkAsDelivered() {
	now := time.Now()
	n.Status = StatusDelivered
	n.DeliveredAt = &now
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.Status = StatusRead
	n.ReadAt = &now

	// Also mark as delivered if not already
	if n.DeliveredAt == nil {
		n.DeliveredAt = &now
	}
}

// MarkAsFailed marks the notification as failed
func (n *Notification) MarkAsFailed(reason string) {
	n.Status = StatusFailed
	n.FailureReason = reason
}

// IsPending checks if the notification is pending
func (n *Notification) IsPending() bool {
	return n.Status == StatusPending
}

// IsRead checks if the notification is read
func (n *Notification) IsRead() bool {
	return n.Status == StatusRead
}

// NotificationInfo represents notification information
type NotificationInfo struct {
	ID            uint64             `json:"id"`
	MemberID      uint64             `json:"memberId"`
	Type          NotificationType   `json:"type"`
	Title         string             `json:"title"`
	Body          string             `json:"body"`
	Data          string             `json:"data,omitempty"`
	Status        NotificationStatus `json:"status"`
	SentAt        *time.Time         `json:"sentAt,omitempty"`
	DeliveredAt   *time.Time         `json:"deliveredAt,omitempty"`
	ReadAt        *time.Time         `json:"readAt,omitempty"`
	FailureReason string             `json:"failureReason,omitempty"`
	CreatedAt     time.Time          `json:"createdAt"`
}

// ToInfo converts Notification to NotificationInfo
func (n *Notification) ToInfo() *NotificationInfo {
	return &NotificationInfo{
		ID:            n.ID,
		MemberID:      n.MemberID,
		Type:          n.Type,
		Title:         n.Title,
		Body:          n.Body,
		Data:          n.Data,
		Status:        n.Status,
		SentAt:        n.SentAt,
		DeliveredAt:   n.DeliveredAt,
		ReadAt:        n.ReadAt,
		FailureReason: n.FailureReason,
		CreatedAt:     n.CreatedAt,
	}
}

// NotificationSummary represents a notification summary
type NotificationSummary struct {
	ID        uint64             `json:"id"`
	Type      NotificationType   `json:"type"`
	Title     string             `json:"title"`
	Status    NotificationStatus `json:"status"`
	CreatedAt time.Time          `json:"createdAt"`
	IsRead    bool               `json:"isRead"`
}

// ToSummary converts Notification to NotificationSummary
func (n *Notification) ToSummary() *NotificationSummary {
	return &NotificationSummary{
		ID:        n.ID,
		Type:      n.Type,
		Title:     n.Title,
		Status:    n.Status,
		CreatedAt: n.CreatedAt,
		IsRead:    n.IsRead(),
	}
}

// Domain errors
var (
	ErrNotificationNotFound    = errors.New("notification not found")
	ErrInvalidMemberID         = errors.New("invalid member ID")
	ErrInvalidTitle            = errors.New("title must be between 1 and 100 characters")
	ErrInvalidBody             = errors.New("body must be between 1 and 500 characters")
	ErrInvalidNotificationType = errors.New("invalid notification type")
	ErrNotificationAlreadySent = errors.New("notification already sent")
	ErrNotificationFailed      = errors.New("notification failed to send")
)
