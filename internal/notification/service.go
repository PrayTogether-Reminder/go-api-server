package notification

import (
	"context"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

type NotificationService struct {
	repo *NotificationRepository
}

func NewNotificationService(repo *NotificationRepository) *NotificationService {
	return &NotificationService{
		repo: repo,
	}
}

// CreatePrayerCompletionNotifications creates in-app notification history for prayer completion
// Excludes the sender from notification recipients (sender doesn't need to see their own notification)
func (s *NotificationService) CreatePrayerCompletionNotifications(
	ctx context.Context,
	db *gorm.DB,
	senderID int64,
	recipientIDs []int64,
	message string,
	prayerTitleID int64,
) error {
	notifications := make([]*model.PrayerCompletionNotification, 0, len(recipientIDs))

	for _, recipientID := range recipientIDs {
		if recipientID == senderID {
			continue
		}

		notification := model.NewPrayerCompletionNotification(
			senderID,
			recipientID,
			prayerTitleID,
			message,
		)
		notifications = append(notifications, notification)
	}

	return s.repo.CreatePrayerCompletionNotifications(ctx, db, notifications)
}
