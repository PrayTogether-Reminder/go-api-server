package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"pray-together/internal/domains/invitation/domain"
)

// GormRepository implements invitation domain repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM repository
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{
		db: db,
	}
}

// Create creates a new invitation
func (r *GormRepository) Create(ctx context.Context, invitation *domain.Invitation) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

// FindByID finds an invitation by ID
func (r *GormRepository) FindByID(ctx context.Context, id uint64) (*domain.Invitation, error) {
	var invitation domain.Invitation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&invitation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &invitation, err
}

// Update updates an invitation
func (r *GormRepository) Update(ctx context.Context, invitation *domain.Invitation) error {
	return r.db.WithContext(ctx).Save(invitation).Error
}

// Delete deletes an invitation
func (r *GormRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Invitation{}, id).Error
}

// FindByInviteeID finds invitations by invitee ID
func (r *GormRepository) FindByInviteeID(ctx context.Context, inviteeID uint64) ([]*domain.Invitation, error) {
	var invitations []*domain.Invitation
	err := r.db.WithContext(ctx).
		Where("invitee_id = ?", inviteeID).
		Order("created_at ASC"). // Match Java: ORDER BY i.createdTime ASC
		Find(&invitations).Error

	return invitations, err
}

// FindByRoomID finds invitations by room ID
func (r *GormRepository) FindByRoomID(ctx context.Context, roomID uint64) ([]*domain.Invitation, error) {
	var invitations []*domain.Invitation
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Find(&invitations).Error

	return invitations, err
}

// FindPendingByInviteeID finds pending invitations by invitee ID
func (r *GormRepository) FindPendingByInviteeID(ctx context.Context, inviteeID uint64) ([]*domain.Invitation, error) {
	var invitations []*domain.Invitation
	err := r.db.WithContext(ctx).
		Where("invitee_id = ? AND status = ? AND expires_at > ?",
						inviteeID, domain.StatusPending, time.Now()).
		Order("created_at ASC"). // Match Java: ORDER BY i.createdTime ASC
		Find(&invitations).Error

	return invitations, err
}

// FindByRoomAndInvitee finds an invitation by room and invitee
func (r *GormRepository) FindByRoomAndInvitee(ctx context.Context, roomID, inviteeID uint64) (*domain.Invitation, error) {
	var invitation domain.Invitation
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND invitee_id = ?", roomID, inviteeID).
		Order("created_at DESC").
		First(&invitation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &invitation, err
}

// MarkExpiredInvitations marks expired invitations
func (r *GormRepository) MarkExpiredInvitations(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Model(&domain.Invitation{}).
		Where("status = ? AND expires_at < ?", domain.StatusPending, before).
		Update("status", domain.StatusExpired).Error
}
