package notification

import (
	"context"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository struct{}

func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{}
}

// CreatePrayerCompletionNotifications creates multiple prayer completion notification records
// Uses batch insert for efficiency
func (r *NotificationRepository) CreatePrayerCompletionNotifications(
	ctx context.Context,
	db *gorm.DB,
	notifications []*model.PrayerCompletionNotification,
) error {
	if len(notifications) == 0 {
		return nil
	}
	return db.WithContext(ctx).Create(&notifications).Error
}
