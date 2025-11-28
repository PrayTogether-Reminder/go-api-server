package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// InvitationStatus represents the status of an invitation
type InvitationStatus string

// Value implements driver.Valuer so InvitationStatus can be stored via database/sql.
func (s InvitationStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan implements sql.Scanner to read InvitationStatus from DB rows.
func (s *InvitationStatus) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*s = InvitationStatus(v)
	case []byte:
		*s = InvitationStatus(string(v))
	case nil:
		*s = ""
	default:
		return fmt.Errorf("unsupported InvitationStatus scan type: %T", value)
	}
	return nil
}

const (
	InvitationPending  InvitationStatus = "PENDING"
	InvitationAccepted InvitationStatus = "ACCEPTED"
	InvitationRejected InvitationStatus = "REJECTED"
)

// Invitation represents an invitation to join a room
type Invitation struct {
	ID           int64            `gorm:"column:id;primaryKey;autoIncrement"`
	InviteeID    int64            `gorm:"column:invitee_id;not null;index:idx_invitation_invitee_status"`
	RoomID       int64            `gorm:"column:room_id;not null;index:idx_invitation_room_invitee_status"`
	InviterName  string           `gorm:"column:inviter_name;type:VARCHAR2(10);not null"`
	Status       InvitationStatus `gorm:"column:status;type:VARCHAR2(20);not null;default:'PENDING';index:idx_invitation_invitee_status,idx_invitation_room_invitee_status"`
	ResponseTime *time.Time       `gorm:"column:response_time"`

	// Relations
	Invitee *Member `gorm:"foreignKey:InviteeID;references:ID;constraint:OnDelete:CASCADE"`
	Room    *Room   `gorm:"foreignKey:RoomID;references:ID;constraint:OnDelete:CASCADE"`

	BaseEntity
}

// TableName specifies the table name for Invitation
func (*Invitation) TableName() string {
	return "invitation"
}

func (i *Invitation) Accept() error {
	if err := i.validatePendingStatus(); err != nil {
		return err
	}

	now := time.Now()
	i.Status = InvitationAccepted
	i.ResponseTime = &now
	return nil
}

func (i *Invitation) Reject() error {
	if err := i.validatePendingStatus(); err != nil {
		return err
	}

	now := time.Now()
	i.Status = InvitationRejected
	i.ResponseTime = &now
	return nil
}

func (i *Invitation) validatePendingStatus() error {
	if i.Status != InvitationPending {
		return fmt.Errorf("이미 응답한 초대장 입니다: inviationId=%d, status=%s", i.ID, i.Status)
	}
	return nil
}
