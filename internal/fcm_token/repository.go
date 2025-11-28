package fcm_token

import (
	"context"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

// FcmTokenRepository encapsulates all persistence logic for FCM tokens.
type FcmTokenRepository struct{}

// NewFcmTokenRepository returns a new repository instance.
func NewFcmTokenRepository() *FcmTokenRepository {
	return &FcmTokenRepository{}
}

// Create inserts a new FCM token row.
func (r *FcmTokenRepository) Create(ctx context.Context, db *gorm.DB, token *model.FcmToken) error {
	return db.WithContext(ctx).Create(token).Error
}

// DeleteByMemberID removes all tokens that belong to the given member.
func (r *FcmTokenRepository) DeleteByMemberID(ctx context.Context, db *gorm.DB, memberID int64) error {
	return db.WithContext(ctx).
		Where("member_id = ?", memberID).
		Delete(&model.FcmToken{}).Error
}

// DeleteByTokenAndMemberID removes a token only if both token value and owner match.
func (r *FcmTokenRepository) DeleteByTokenAndMemberID(ctx context.Context, db *gorm.DB, token string, memberID int64) error {
	return db.WithContext(ctx).
		Where("member_id = ? AND token = ?", memberID, token).
		Delete(&model.FcmToken{}).Error
}
