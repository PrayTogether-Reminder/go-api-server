package application

import (
	"context"
	"time"

	"pray-together/internal/domains/notification/domain"
)

// SendBulkNotificationRequest represents the request to send bulk notifications
type SendBulkNotificationRequest struct {
	RecipientIDs []uint64
	Title        string
	Body         string
	Type         string
	Data         map[string]interface{}
}

// BulkNotificationResult represents the result of bulk notification sending
type BulkNotificationResult struct {
	SuccessCount int      `json:"successCount"`
	FailureCount int      `json:"failureCount"`
	FailedIDs    []uint64 `json:"failedIds,omitempty"`
}

// SendBulkNotificationUseCase handles sending bulk notifications
type SendBulkNotificationUseCase struct {
	notificationService  *domain.Service
	sendPushNotification func(ctx context.Context, memberID uint64, title, body string, data map[string]interface{}) error
}

// NewSendBulkNotificationUseCase creates a new send bulk notification use case
func NewSendBulkNotificationUseCase(
	notificationService *domain.Service,
	sendPushNotification func(ctx context.Context, memberID uint64, title, body string, data map[string]interface{}) error,
) *SendBulkNotificationUseCase {
	return &SendBulkNotificationUseCase{
		notificationService:  notificationService,
		sendPushNotification: sendPushNotification,
	}
}

// Execute sends bulk notifications
func (u *SendBulkNotificationUseCase) Execute(ctx context.Context, req *SendBulkNotificationRequest) *BulkNotificationResult {
	result := &BulkNotificationResult{
		FailedIDs: make([]uint64, 0),
	}

	for _, recipientID := range req.RecipientIDs {
		// Create notification record
		notification := &domain.Notification{
			RecipientID: recipientID,
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
			result.FailureCount++
			result.FailedIDs = append(result.FailedIDs, recipientID)
			continue
		}

		// Send push notification
		if u.sendPushNotification != nil {
			if err := u.sendPushNotification(ctx, recipientID, req.Title, req.Body, req.Data); err != nil {
				// Log error but don't fail the whole operation
				continue
			}
		}

		result.SuccessCount++
	}

	return result
}
