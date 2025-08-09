package domain

import (
	"context"
	"encoding/json"
	"time"
)

// Service represents notification domain service
type Service struct {
	repo      Repository
	getTokens func(ctx context.Context, memberID uint64) ([]string, error)
	sendPush  func(ctx context.Context, tokens []string, title, body string, data map[string]interface{}) error
}

// NewService creates a new notification service
func NewService(
	repo Repository,
	getTokens func(ctx context.Context, memberID uint64) ([]string, error),
	sendPush func(ctx context.Context, tokens []string, title, body string, data map[string]interface{}) error,
) *Service {
	return &Service{
		repo:      repo,
		getTokens: getTokens,
		sendPush:  sendPush,
	}
}

// CreateNotification creates and sends a notification
func (s *Service) CreateNotification(ctx context.Context, memberID uint64, notifType NotificationType, title, body string, data map[string]interface{}) (*Notification, error) {
	// Convert data to JSON string
	dataJSON := ""
	if data != nil {
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		dataJSON = string(jsonBytes)
	}

	// Create notification
	notification, err := NewNotification(memberID, notifType, title, body, dataJSON)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, err
	}

	// Send push notification
	if s.getTokens != nil && s.sendPush != nil {
		tokens, err := s.getTokens(ctx, memberID)
		if err == nil && len(tokens) > 0 {
			if err := s.sendPush(ctx, tokens, title, body, data); err != nil {
				notification.MarkAsFailed(err.Error())
			} else {
				notification.MarkAsSent()
			}

			// Update notification status
			_ = s.repo.Update(ctx, notification)
		}
	}

	return notification, nil
}

// GetNotification gets a notification by ID
func (s *Service) GetNotification(ctx context.Context, id uint64) (*Notification, error) {
	notification, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if notification == nil {
		return nil, ErrNotificationNotFound
	}

	return notification, nil
}

// GetMemberNotifications gets notifications for a member
func (s *Service) GetMemberNotifications(ctx context.Context, memberID uint64, limit, offset int) ([]*Notification, error) {
	return s.repo.FindByMemberID(ctx, memberID, limit, offset)
}

// GetUnreadNotifications gets unread notifications for a member
func (s *Service) GetUnreadNotifications(ctx context.Context, memberID uint64) ([]*Notification, error) {
	return s.repo.FindUnreadByMemberID(ctx, memberID)
}

// MarkAsRead marks a notification as read
func (s *Service) MarkAsRead(ctx context.Context, notificationID uint64, memberID uint64) error {
	notification, err := s.repo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}

	if notification == nil {
		return ErrNotificationNotFound
	}

	// Verify member owns the notification
	if notification.MemberID != memberID {
		return ErrNotificationNotFound
	}

	notification.MarkAsRead()
	return s.repo.Update(ctx, notification)
}

// MarkAllAsRead marks all notifications as read for a member
func (s *Service) MarkAllAsRead(ctx context.Context, memberID uint64) error {
	return s.repo.MarkAllAsRead(ctx, memberID)
}

// GetUnreadCount gets the count of unread notifications
func (s *Service) GetUnreadCount(ctx context.Context, memberID uint64) (int, error) {
	return s.repo.CountUnreadByMemberID(ctx, memberID)
}

// SendPrayerNotification sends a prayer-related notification
func (s *Service) SendPrayerNotification(ctx context.Context, memberID uint64, roomName, prayerContent string) (*Notification, error) {
	title := "New Prayer in " + roomName
	body := prayerContent
	if len(body) > 100 {
		body = body[:100] + "..."
	}

	data := map[string]interface{}{
		"type":     "prayer",
		"roomName": roomName,
	}

	return s.CreateNotification(ctx, memberID, NotificationTypePrayer, title, body, data)
}

// SendInvitationNotification sends an invitation notification
func (s *Service) SendInvitationNotification(ctx context.Context, memberID uint64, inviterName, roomName string) (*Notification, error) {
	title := "Room Invitation"
	body := inviterName + " invited you to join " + roomName

	data := map[string]interface{}{
		"type":        "invitation",
		"inviterName": inviterName,
		"roomName":    roomName,
	}

	return s.CreateNotification(ctx, memberID, NotificationTypeInvitation, title, body, data)
}

// ProcessPendingNotifications processes pending notifications
func (s *Service) ProcessPendingNotifications(ctx context.Context) error {
	notifications, err := s.repo.FindPendingNotifications(ctx, 100)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		// Try to send the notification
		if s.getTokens != nil && s.sendPush != nil {
			tokens, err := s.getTokens(ctx, notification.MemberID)
			if err == nil && len(tokens) > 0 {
				data := make(map[string]interface{})
				if notification.Data != "" {
					_ = json.Unmarshal([]byte(notification.Data), &data)
				}

				if err := s.sendPush(ctx, tokens, notification.Title, notification.Body, data); err != nil {
					notification.MarkAsFailed(err.Error())
				} else {
					notification.MarkAsSent()
				}

				_ = s.repo.Update(ctx, notification)
			}
		}
	}

	return nil
}

// CleanupOldNotifications deletes old notifications
func (s *Service) CleanupOldNotifications(ctx context.Context) error {
	// Delete notifications older than 30 days
	before := time.Now().AddDate(0, 0, -30)
	return s.repo.DeleteOldNotifications(ctx, before)
}
