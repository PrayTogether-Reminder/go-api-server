package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"pray-together/internal/domains/notification/domain"
)

// GormRepository implements domain.Repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM repository
func NewGormRepository(db *gorm.DB) domain.Repository {
	return &GormRepository{db: db}
}

// Create creates a new notification
func (r *GormRepository) Create(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// Update updates a notification
func (r *GormRepository) Update(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

// Delete deletes a notification
func (r *GormRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Notification{}, id).Error
}

// FindByID finds a notification by ID
func (r *GormRepository) FindByID(ctx context.Context, id uint64) (*domain.Notification, error) {
	var notification domain.Notification
	err := r.db.WithContext(ctx).First(&notification, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, err
	}
	return &notification, nil
}

// FindByMemberID finds notifications by member ID
func (r *GormRepository) FindByMemberID(ctx context.Context, memberID uint64, limit, offset int) ([]*domain.Notification, error) {
	var notifications []*domain.Notification

	query := r.db.WithContext(ctx).
		Where("member_id = ?", memberID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&notifications).Error
	return notifications, err
}

// FindUnreadByMemberID finds unread notifications by member ID
func (r *GormRepository) FindUnreadByMemberID(ctx context.Context, memberID uint64) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND is_read = ?", memberID, false).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// FindByType finds notifications by type for a member
func (r *GormRepository) FindByType(ctx context.Context, memberID uint64, notifType domain.NotificationType) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND type = ?", memberID, notifType).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// FindPendingNotifications finds pending notifications
func (r *GormRepository) FindPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	query := r.db.WithContext(ctx).
		Where("status = ?", "PENDING").
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&notifications).Error
	return notifications, err
}

// MarkAsRead marks a notification as read
func (r *GormRepository) MarkAsRead(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ?", id).
		Update("is_read", true).Error
}

// MarkAllAsRead marks all notifications as read for a member
func (r *GormRepository) MarkAllAsRead(ctx context.Context, memberID uint64) error {
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("member_id = ? AND is_read = ?", memberID, false).
		Update("is_read", true).Error
}

// CountUnreadByMemberID counts unread notifications for a member
func (r *GormRepository) CountUnreadByMemberID(ctx context.Context, memberID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("member_id = ? AND is_read = ?", memberID, false).
		Count(&count).Error
	return int(count), err
}

// DeleteOldNotifications deletes notifications older than specified date
func (r *GormRepository) DeleteOldNotifications(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&domain.Notification{}).Error
}
