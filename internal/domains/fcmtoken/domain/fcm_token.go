package domain

import (
	"errors"
	"time"
)

// DeviceType represents the type of device
type DeviceType string

const (
	DeviceTypeIOS     DeviceType = "IOS"
	DeviceTypeAndroid DeviceType = "ANDROID"
	DeviceTypeWeb     DeviceType = "WEB"
)

// BaseEntity contains common fields for all entities in fcmtoken domain
type BaseEntity struct {
	CreatedAt time.Time  `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`
}

// FCMToken represents an FCM token entity
type FCMToken struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	MemberID   uint64     `gorm:"column:member_id;not null;index" json:"memberId"`
	Token      string     `gorm:"column:token;uniqueIndex;not null" json:"token"`
	DeviceType DeviceType `gorm:"column:device_type;not null" json:"deviceType"`
	DeviceID   string     `gorm:"column:device_id;index" json:"deviceId,omitempty"`
	IsActive   bool       `gorm:"column:is_active;default:true" json:"isActive"`
	LastUsedAt time.Time  `gorm:"column:last_used_at" json:"lastUsedAt"`
	BaseEntity
}

// TableName specifies the table name for FCMToken
func (FCMToken) TableName() string {
	return "fcm_token"
}

// NewFCMToken creates a new FCM token
func NewFCMToken(memberID uint64, token string, deviceType DeviceType, deviceID string) (*FCMToken, error) {
	fcmToken := &FCMToken{
		MemberID:   memberID,
		Token:      token,
		DeviceType: deviceType,
		DeviceID:   deviceID,
		IsActive:   true,
		LastUsedAt: time.Now(),
	}

	if err := fcmToken.Validate(); err != nil {
		return nil, err
	}

	return fcmToken, nil
}

// Validate validates FCM token data
func (f *FCMToken) Validate() error {
	if f.MemberID == 0 {
		return ErrInvalidMemberID
	}

	if f.Token == "" || len(f.Token) > 500 {
		return ErrInvalidToken
	}

	if f.DeviceType != DeviceTypeIOS && f.DeviceType != DeviceTypeAndroid && f.DeviceType != DeviceTypeWeb {
		return ErrInvalidDeviceType
	}

	return nil
}

// Activate activates the token
func (f *FCMToken) Activate() {
	f.IsActive = true
	f.LastUsedAt = time.Now()
}

// Deactivate deactivates the token
func (f *FCMToken) Deactivate() {
	f.IsActive = false
}

// UpdateLastUsed updates the last used timestamp
func (f *FCMToken) UpdateLastUsed() {
	f.LastUsedAt = time.Now()
}

// IsStale checks if the token is stale (not used for more than 30 days)
func (f *FCMToken) IsStale() bool {
	return time.Since(f.LastUsedAt) > 30*24*time.Hour
}

// FCMTokenInfo represents FCM token information
type FCMTokenInfo struct {
	ID         uint64     `json:"id"`
	MemberID   uint64     `json:"memberId"`
	Token      string     `json:"token"`
	DeviceType DeviceType `json:"deviceType"`
	DeviceID   string     `json:"deviceId,omitempty"`
	IsActive   bool       `json:"isActive"`
	LastUsedAt time.Time  `json:"lastUsedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// ToInfo converts FCMToken to FCMTokenInfo
func (f *FCMToken) ToInfo() *FCMTokenInfo {
	return &FCMTokenInfo{
		ID:         f.ID,
		MemberID:   f.MemberID,
		Token:      f.Token,
		DeviceType: f.DeviceType,
		DeviceID:   f.DeviceID,
		IsActive:   f.IsActive,
		LastUsedAt: f.LastUsedAt,
		CreatedAt:  f.CreatedAt,
	}
}

// Domain errors
var (
	ErrFCMTokenNotFound   = errors.New("FCM token not found")
	ErrInvalidMemberID    = errors.New("invalid member ID")
	ErrInvalidToken       = errors.New("invalid FCM token")
	ErrInvalidDeviceType  = errors.New("invalid device type")
	ErrTokenAlreadyExists = errors.New("FCM token already exists")
	ErrTokenInactive      = errors.New("FCM token is inactive")
	ErrNotAuthorized      = errors.New("not authorized")
)
