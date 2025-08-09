package application

import (
	"context"
	"time"

	"pray-together/internal/domains/notification/domain"
)

// SendNotificationRequest represents the request to send a notification
type SendNotificationRequest struct {
	RecipientID uint64
	Title       string
	Body        string
	Type        string
	Data        map[string]interface{}
}

// SendNotificationUseCase handles sending notifications
type SendNotificationUseCase struct {
	notificationService  *domain.Service
	sendPushNotification func(ctx context.Context, memberID uint64, title, body string, data map[string]interface{}) error
}

// NewSendNotificationUseCase creates a new send notification use case
func NewSendNotificationUseCase(
	notificationService *domain.Service,
	sendPushNotification func(ctx context.Context, memberID uint64, title, body string, data map[string]interface{}) error,
) *SendNotificationUseCase {
	return &SendNotificationUseCase{
		notificationService:  notificationService,
		sendPushNotification: sendPushNotification,
	}
}

// Execute sends a notification
func (u *SendNotificationUseCase) Execute(ctx context.Context, req *SendNotificationRequest) error {
	// Create notification record
	notification := &domain.Notification{
		RecipientID: req.RecipientID,
		Title:       req.Title,
		Body:        req.Body,
		Type:        domain.NotificationType(req.Type),
		IsRead:      false,
		BaseEntity: domain.BaseEntity{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Save notification to database
	if err := u.notificationService.CreateNotification(ctx, notification); err != nil {
		return err
	}

	// Send push notification
	if u.sendPushNotification != nil {
		_ = u.sendPushNotification(ctx, req.RecipientID, req.Title, req.Body, req.Data)
	}

	return nil
}
