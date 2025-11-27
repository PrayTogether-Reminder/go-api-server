package model

import (
	"time"
)

// InvitationStatus represents the status of an invitation
// Java의 InvitationStatus enum과 동일
type InvitationStatus string

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

// validatePendingStatus checks if the invitation is in PENDING status
// Java: private void validatePendingStatus() throws AlreadyRespondedInvitationException
//func (i *Invitation) validatePendingStatus() error {
//	if i.Status != InvitationPending {
//		// Java에서는 AlreadyRespondedInvitationException(this.id, this.status)를 던짐
//		return fmt.Errorf("already responded invitation: id=%d, status=%s", i.ID, i.Status)
//	}
//	return nil
//}
