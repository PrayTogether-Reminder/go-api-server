package domain

import (
	"errors"
	"time"
)

// InvitationStatus represents the status of an invitation
type InvitationStatus string

const (
	StatusPending  InvitationStatus = "PENDING"
	StatusAccepted InvitationStatus = "ACCEPTED"
	StatusRejected InvitationStatus = "REJECTED"
	StatusExpired  InvitationStatus = "EXPIRED"
)

// BaseEntity contains common fields for all entities in invitation domain
type BaseEntity struct {
	CreatedAt time.Time  `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`
}

// Invitation represents an invitation entity
type Invitation struct {
	ID           uint64           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RoomID       uint64           `gorm:"column:room_id;not null;index" json:"roomId"`
	InviterID    uint64           `gorm:"column:inviter_id;not null;index" json:"inviterId"`
	InviterName  string           `gorm:"column:inviter_name;not null" json:"inviterName"` // Added to match Java
	InviteeID    uint64           `gorm:"column:invitee_id;not null;index" json:"inviteeId"`
	Status       InvitationStatus `gorm:"column:status;not null;default:'PENDING'" json:"status"`
	Message      string           `gorm:"column:message;type:text" json:"message,omitempty"`
	ExpiresAt    time.Time        `gorm:"column:expires_at;not null" json:"expiresAt"`
	ResponseTime *time.Time       `gorm:"column:response_time" json:"responseTime,omitempty"` // Renamed to match Java
	BaseEntity
}

// TableName specifies the table name for Invitation
func (Invitation) TableName() string {
	return "invitation"
}

// NewInvitation creates a new invitation
func NewInvitation(roomID, inviterID, inviteeID uint64, message string, expiresAt time.Time) (*Invitation, error) {
	invitation := &Invitation{
		RoomID:    roomID,
		InviterID: inviterID,
		InviteeID: inviteeID,
		Status:    StatusPending,
		Message:   message,
		ExpiresAt: expiresAt,
	}

	if err := invitation.Validate(); err != nil {
		return nil, err
	}

	return invitation, nil
}

// Validate validates invitation data
func (i *Invitation) Validate() error {
	if i.RoomID == 0 {
		return ErrInvalidRoomID
	}

	if i.InviterID == 0 {
		return ErrInvalidInviterID
	}

	if i.InviteeID == 0 {
		return ErrInvalidInviteeID
	}

	if i.InviterID == i.InviteeID {
		return ErrSelfInvitation
	}

	if len(i.Message) > 500 {
		return ErrMessageTooLong
	}

	return nil
}

// IsExpired checks if the invitation is expired
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsPending checks if the invitation is pending
func (i *Invitation) IsPending() bool {
	return i.Status == StatusPending && !i.IsExpired()
}

// CanRespond checks if the invitation can be responded to
func (i *Invitation) CanRespond() bool {
	return i.Status == StatusPending && !i.IsExpired()
}

// Accept accepts the invitation
func (i *Invitation) Accept() error {
	if !i.CanRespond() {
		if i.IsExpired() {
			return ErrInvitationExpired
		}
		return ErrAlreadyResponded
	}

	now := time.Now()
	i.Status = StatusAccepted
	i.ResponseTime = &now
	return nil
}

// Reject rejects the invitation
func (i *Invitation) Reject() error {
	if !i.CanRespond() {
		if i.IsExpired() {
			return ErrInvitationExpired
		}
		return ErrAlreadyResponded
	}

	now := time.Now()
	i.Status = StatusRejected
	i.ResponseTime = &now
	return nil
}

// MarkAsExpired marks the invitation as expired
func (i *Invitation) MarkAsExpired() {
	i.Status = StatusExpired
}

// InvitationInfo represents invitation information (matching Java InvitationInfo)
type InvitationInfo struct {
	InvitationID    uint64    `json:"invitationId"`
	InviterName     string    `json:"inviterName"`
	RoomName        string    `json:"roomName"`
	RoomDescription string    `json:"roomDescription"`
	CreatedTime     time.Time `json:"createdTime"`
}

// ToInfo converts Invitation to InvitationInfo
func (i *Invitation) ToInfo() *InvitationInfo {
	return &InvitationInfo{
		InvitationID: i.ID,
		InviterName:  i.InviterName,
		CreatedTime:  i.CreatedAt,
		// RoomName and RoomDescription need to be filled from room service
	}
}

// Domain errors
var (
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrInvitationExpired  = errors.New("invitation has expired")
	ErrAlreadyResponded   = errors.New("invitation already responded")
	ErrInvalidRoomID      = errors.New("invalid room ID")
	ErrInvalidInviterID   = errors.New("invalid inviter ID")
	ErrInvalidInviteeID   = errors.New("invalid invitee ID")
	ErrSelfInvitation     = errors.New("cannot invite yourself")
	ErrMessageTooLong     = errors.New("invitation message too long")
	ErrAlreadyInvited     = errors.New("already invited to this room")
	ErrNotAuthorized      = errors.New("not authorized to send invitation")
)
