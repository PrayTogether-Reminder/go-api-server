package application

import (
	"context"

	"pray-together/internal/domains/notification/domain"
)

// MarkNotificationReadRequest represents the request to mark a notification as read
type MarkNotificationReadRequest struct {
	NotificationID uint64
	MemberID       uint64
}

// MarkNotificationReadUseCase handles marking notifications as read
type MarkNotificationReadUseCase struct {
	notificationService *domain.Service
}

// NewMarkNotificationReadUseCase creates a new mark notification read use case
func NewMarkNotificationReadUseCase(notificationService *domain.Service) *MarkNotificationReadUseCase {
	return &MarkNotificationReadUseCase{
		notificationService: notificationService,
	}
}

// Execute marks a notification as read
func (u *MarkNotificationReadUseCase) Execute(ctx context.Context, req *MarkNotificationReadRequest) error {
	return u.notificationService.MarkAsRead(ctx, req.NotificationID, req.MemberID)
}

// MarkAllAsRead marks all notifications as read for a user
func (u *MarkNotificationReadUseCase) MarkAllAsRead(ctx context.Context, memberID uint64) error {
	return u.notificationService.MarkAllAsRead(ctx, memberID)
}
