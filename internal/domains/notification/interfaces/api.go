package interfaces

import (
	"context"
	"pray-together/internal/domains/notification/domain"
)

// API represents the public interface for notification domain
type API interface {
	// Notification operations
	SendPrayerNotification(ctx context.Context, memberID uint64, roomName, prayerContent string) error
	SendInvitationNotification(ctx context.Context, memberID uint64, inviterName, roomName string) error
	GetUnreadCount(ctx context.Context, memberID uint64) (int, error)
	MarkAsRead(ctx context.Context, notificationID uint64, memberID uint64) error
}

// Module represents the notification module with all its components
type Module struct {
	Service          *domain.Service
	Repository       domain.Repository
	SendNotification func(ctx context.Context, memberID uint64, message string, notificationType string) error
}

// NewModule creates a new notification module
func NewModule(
	repo domain.Repository,
	getTokens func(ctx context.Context, memberID uint64) ([]string, error),
	sendPush func(ctx context.Context, tokens []string, title, body string, data map[string]interface{}) error,
) *Module {
	return &Module{
		Service:    domain.NewService(repo, getTokens, sendPush),
		Repository: repo,
	}
}

// SendPrayerNotification implements API interface
func (m *Module) SendPrayerNotification(ctx context.Context, memberID uint64, roomName, prayerContent string) error {
	_, err := m.Service.SendPrayerNotification(ctx, memberID, roomName, prayerContent)
	return err
}

// SendInvitationNotification implements API interface
func (m *Module) SendInvitationNotification(ctx context.Context, memberID uint64, inviterName, roomName string) error {
	_, err := m.Service.SendInvitationNotification(ctx, memberID, inviterName, roomName)
	return err
}

// GetUnreadCount implements API interface
func (m *Module) GetUnreadCount(ctx context.Context, memberID uint64) (int, error) {
	return m.Service.GetUnreadCount(ctx, memberID)
}

// MarkAsRead implements API interface
func (m *Module) MarkAsRead(ctx context.Context, notificationID uint64, memberID uint64) error {
	return m.Service.MarkAsRead(ctx, notificationID, memberID)
}
