package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"pray-together/internal/domains/fcmtoken/domain"
)

// GormRepository implements domain.Repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM repository
func NewGormRepository(db *gorm.DB) domain.Repository {
	return &GormRepository{db: db}
}

// Create creates a new FCM token
func (r *GormRepository) Create(ctx context.Context, token *domain.FCMToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// Update updates an FCM token
func (r *GormRepository) Update(ctx context.Context, token *domain.FCMToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

// Delete deletes an FCM token
func (r *GormRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.FCMToken{}, id).Error
}

// FindByID finds an FCM token by ID
func (r *GormRepository) FindByID(ctx context.Context, id uint64) (*domain.FCMToken, error) {
	var token domain.FCMToken
	err := r.db.WithContext(ctx).First(&token, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrFCMTokenNotFound
		}
		return nil, err
	}
	return &token, nil
}

// FindByMemberID finds FCM tokens by member ID
func (r *GormRepository) FindByMemberID(ctx context.Context, memberID uint64) ([]*domain.FCMToken, error) {
	var tokens []*domain.FCMToken
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND is_active = ?", memberID, true).
		Find(&tokens).Error
	return tokens, err
}

// FindByToken finds an FCM token by token value
func (r *GormRepository) FindByToken(ctx context.Context, token string) (*domain.FCMToken, error) {
	var fcmToken domain.FCMToken
	err := r.db.WithContext(ctx).
		Where("token = ?", token).
		First(&fcmToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrFCMTokenNotFound
		}
		return nil, err
	}
	return &fcmToken, nil
}

// FindByDeviceID finds FCM tokens by device ID
func (r *GormRepository) FindByDeviceID(ctx context.Context, memberID uint64, deviceID string) (*domain.FCMToken, error) {
	var token domain.FCMToken
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND device_id = ?", memberID, deviceID).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrFCMTokenNotFound
		}
		return nil, err
	}
	return &token, nil
}

// DeleteByToken deletes an FCM token by token value
func (r *GormRepository) DeleteByToken(ctx context.Context, memberID uint64, token string) error {
	result := r.db.WithContext(ctx).
		Where("member_id = ? AND token = ?", memberID, token).
		Delete(&domain.FCMToken{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrFCMTokenNotFound
	}
	return nil
}

// DeleteByDeviceID deletes FCM tokens by device ID
func (r *GormRepository) DeleteByDeviceID(ctx context.Context, memberID uint64, deviceID string) error {
	result := r.db.WithContext(ctx).
		Where("member_id = ? AND device_id = ?", memberID, deviceID).
		Delete(&domain.FCMToken{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrFCMTokenNotFound
	}
	return nil
}

// DeactivateExpiredTokens deactivates expired tokens
func (r *GormRepository) DeactivateExpiredTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&domain.FCMToken{}).
		Where("is_active = ? AND updated_at < NOW() - INTERVAL '30 days'", true).
		Update("is_active", false).Error
}

// FindActiveByMemberID finds active FCM tokens by member ID
func (r *GormRepository) FindActiveByMemberID(ctx context.Context, memberID uint64) ([]*domain.FCMToken, error) {
	var tokens []*domain.FCMToken
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND is_active = ?", memberID, true).
		Find(&tokens).Error
	return tokens, err
}

// DeactivateByMemberID deactivates all tokens for a member
func (r *GormRepository) DeactivateByMemberID(ctx context.Context, memberID uint64) error {
	return r.db.WithContext(ctx).
		Model(&domain.FCMToken{}).
		Where("member_id = ?", memberID).
		Update("is_active", false).Error
}

// DeleteByMemberID deletes all tokens for a member (for Java-style registration)
func (r *GormRepository) DeleteByMemberID(ctx context.Context, memberID uint64) error {
	return r.db.WithContext(ctx).
		Where("member_id = ?", memberID).
		Delete(&domain.FCMToken{}).Error
}

// DeleteStaleTokens deletes tokens older than specified days
func (r *GormRepository) DeleteStaleTokens(ctx context.Context, days int) error {
	// SQLite doesn't support INTERVAL, use datetime function instead
	return r.db.WithContext(ctx).
		Where("last_used_at < datetime('now', '-' || ? || ' days')", days).
		Delete(&domain.FCMToken{}).Error
}
