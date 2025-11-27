package invitation

import (
	"context"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

// InvitationInfo represents invitation information with inviter and room details
type InvitationInfo struct {
	InvitationID    int64     `json:"invitationId"`
	InviterName     string    `json:"inviterName"`
	RoomName        string    `json:"roomName"`
	RoomDescription string    `json:"roomDescription"`
	CreatedTime     time.Time `json:"createdTime"`
}

type InvitationRepository struct{}

func NewInvitationRepository() *InvitationRepository {
	return &InvitationRepository{}
}

// CreateAll saves multiple invitations
func (r *InvitationRepository) CreateAll(ctx context.Context, db *gorm.DB, invitations []*model.Invitation) error {
	if len(invitations) == 0 {
		return nil
	}
	return db.WithContext(ctx).Create(&invitations).Error
}

// FindByInviteeIDAndID retrieves an invitation by invitee ID and invitation ID
//func (r *InvitationRepository) FindByInviteeIDAndID(ctx context.Context, db *gorm.DB, inviteeID, invitationID int64) (*model.Invitation, error) {
//	var invitation model.Invitation
//	err := db.WithContext(ctx).
//		Where("invitee_id = ? AND id = ?", inviteeID, invitationID).
//		First(&invitation).Error
//
//	if err != nil {
//		return nil, err
//	}
//	return &invitation, nil
//}

// FindInfosByInviteeIDAndStatus retrieves invitation infos by invitee ID and status
func (r *InvitationRepository) FindInfosByInviteeIDAndStatus(ctx context.Context, db *gorm.DB, inviteeID int64, status model.InvitationStatus) ([]InvitationInfo, error) {
	var results []InvitationInfo

	err := db.WithContext(ctx).
		Table("invitation i").
		Select(`
			i.id as invitation_id,
			i.inviter_name,
			r.name as room_name,
			r.description as room_description,
			i.created_time
		`).
		Joins("JOIN room r ON i.room_id = r.id").
		Where("i.invitee_id = ? AND i.status = ?", inviteeID, status).
		Order("i.created_time DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// FindByRoomIDAndStatusAndInviteeIDs retrieves invitations by room ID, status, and invitee IDs
func (r *InvitationRepository) FindByRoomIDAndStatusAndInviteeIDs(ctx context.Context, db *gorm.DB, roomID int64, status model.InvitationStatus, inviteeIDs []int64) ([]*model.Invitation, error) {
	if len(inviteeIDs) == 0 {
		return []*model.Invitation{}, nil
	}

	var invitations []*model.Invitation
	err := db.WithContext(ctx).
		Where("room_id = ? AND status = ? AND invitee_id IN ?", roomID, status, inviteeIDs).
		Find(&invitations).Error

	if err != nil {
		return nil, err
	}

	return invitations, nil
}

// Update saves the invitation
//func (r *InvitationRepository) Update(ctx context.Context, db *gorm.DB, invitation *model.Invitation) error {
//	return db.WithContext(ctx).Save(invitation).Error
//}
