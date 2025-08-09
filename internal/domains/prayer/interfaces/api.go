package interfaces

import (
	"context"
	"pray-together/internal/domains/prayer/domain"
)

// API represents the public interface for prayer domain
// Other domains should use this interface instead of accessing internal components
type API interface {
	// Prayer operations
	GetPrayerInfo(ctx context.Context, prayerID uint64, requestorID uint64) (*domain.PrayerInfo, error)
	GetMemberPrayerCount(ctx context.Context, memberID uint64) (int, error)
	GetRoomPrayerCount(ctx context.Context, roomID uint64) (int, error)

	// Statistics
	GetMemberPrayerStatistics(ctx context.Context, memberID uint64) (*domain.PrayerStatistics, error)
}

// Module represents the prayer module with all its components
type Module struct {
	Service          *domain.Service
	Repository       domain.Repository
	GetMemberName    func(ctx context.Context, memberID uint64) (string, error)
	GetRoomMemberIDs func(ctx context.Context, roomID uint64) ([]uint64, error)
	SendNotification func(ctx context.Context, senderID uint64, recipientIDs []uint64, message string, prayerTitleID uint64) error
}

// NewModule creates a new prayer module
func NewModule(
	repo domain.Repository,
	validateRoomAccess func(ctx context.Context, roomID, memberID uint64) error,
	recordMemberPray func(ctx context.Context, roomID, memberID uint64) error,
	getMemberName func(ctx context.Context, memberID uint64) (string, error),
	getRoomMemberIDs func(ctx context.Context, roomID uint64) ([]uint64, error),
	sendNotification func(ctx context.Context, senderID uint64, recipientIDs []uint64, message string, prayerTitleID uint64) error,
) *Module {
	return &Module{
		Service:          domain.NewService(repo, validateRoomAccess, recordMemberPray),
		Repository:       repo,
		GetMemberName:    getMemberName,
		GetRoomMemberIDs: getRoomMemberIDs,
		SendNotification: sendNotification,
	}
}

// GetPrayerInfo implements API interface
func (m *Module) GetPrayerInfo(ctx context.Context, prayerID uint64, requestorID uint64) (*domain.PrayerInfo, error) {
	prayer, err := m.Service.GetPrayer(ctx, prayerID, requestorID)
	if err != nil {
		return nil, err
	}
	return prayer.ToInfo(), nil
}

// GetMemberPrayerCount implements API interface
func (m *Module) GetMemberPrayerCount(ctx context.Context, memberID uint64) (int, error) {
	return m.Repository.CountByMemberID(ctx, memberID)
}

// GetRoomPrayerCount implements API interface
func (m *Module) GetRoomPrayerCount(ctx context.Context, roomID uint64) (int, error) {
	return m.Repository.CountByRoomID(ctx, roomID)
}

// GetMemberPrayerStatistics implements API interface
func (m *Module) GetMemberPrayerStatistics(ctx context.Context, memberID uint64) (*domain.PrayerStatistics, error) {
	return m.Service.GetPrayerStatistics(ctx, memberID)
}
