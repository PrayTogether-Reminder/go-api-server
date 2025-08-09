package application

import (
	"context"

	"pray-together/internal/domains/notification/domain"
)

// ListNotificationsRequest represents the request to list notifications
type ListNotificationsRequest struct {
	MemberID   uint64
	Limit      int
	Offset     int
	UnreadOnly bool
}

// ListNotificationsUseCase handles listing notifications
type ListNotificationsUseCase struct {
	notificationService *domain.Service
}

// NewListNotificationsUseCase creates a new list notifications use case
func NewListNotificationsUseCase(notificationService *domain.Service) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{
		notificationService: notificationService,
	}
}

// Execute lists notifications for a user
func (u *ListNotificationsUseCase) Execute(ctx context.Context, req *ListNotificationsRequest) ([]*domain.NotificationInfo, error) {
	notifications, err := u.notificationService.GetNotificationsByRecipient(ctx, req.MemberID, req.Limit, req.Offset, req.UnreadOnly)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.NotificationInfo, len(notifications))
	for i, notification := range notifications {
		result[i] = notification.ToInfo()
	}

	return result, nil
}

// GetUnreadCount gets the count of unread notifications
func (u *ListNotificationsUseCase) GetUnreadCount(ctx context.Context, memberID uint64) (int, error) {
	return u.notificationService.GetUnreadCount(ctx, memberID)
}
