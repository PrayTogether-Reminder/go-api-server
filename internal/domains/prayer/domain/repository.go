package domain

import (
	"context"
	"time"
)

// Repository interface for prayer domain
type Repository interface {
	// Basic CRUD operations (legacy)
	Create(ctx context.Context, prayer *Prayer) error
	FindByID(ctx context.Context, id uint64) (*Prayer, error)
	Update(ctx context.Context, prayer *Prayer) error
	Delete(ctx context.Context, id uint64) error

	// Query operations (legacy)
	FindByMemberID(ctx context.Context, memberID uint64, limit, offset int) ([]*Prayer, error)
	FindByRoomID(ctx context.Context, roomID uint64, limit, offset int) ([]*Prayer, error)
	FindByMemberAndRoom(ctx context.Context, memberID, roomID uint64, limit, offset int) ([]*Prayer, error)
	FindSharedPrayersByRoom(ctx context.Context, roomID uint64, limit, offset int) ([]*Prayer, error)

	// Filtered queries (legacy)
	FindAnsweredPrayers(ctx context.Context, memberID uint64, limit, offset int) ([]*Prayer, error)
	FindUnansweredPrayers(ctx context.Context, memberID uint64, limit, offset int) ([]*Prayer, error)
	FindPrayersByDateRange(ctx context.Context, memberID uint64, startDate, endDate time.Time) ([]*Prayer, error)

	// Count operations (legacy)
	CountByMemberID(ctx context.Context, memberID uint64) (int, error)
	CountByRoomID(ctx context.Context, roomID uint64) (int, error)
	CountAnsweredByMemberID(ctx context.Context, memberID uint64) (int, error)
	CountSharedByRoomID(ctx context.Context, roomID uint64) (int, error)
	CountByRoomIDSince(ctx context.Context, roomID uint64, since time.Time) (int64, error)

	// Existence checks
	ExistsByID(ctx context.Context, id uint64) (bool, error)

	// PrayerTitle operations
	CreatePrayerTitle(ctx context.Context, title *PrayerTitle) error
	FindPrayerTitleByID(ctx context.Context, id uint64) (*PrayerTitle, error)
	FindPrayerTitleWithContents(ctx context.Context, id uint64) (*PrayerTitle, error)
	FindPrayerTitlesByRoomID(ctx context.Context, roomID uint64, after uint64, limit int) ([]*PrayerTitle, error)
	UpdatePrayerTitle(ctx context.Context, title *PrayerTitle) error
	DeletePrayerTitle(ctx context.Context, id uint64) error

	// PrayerTitle infinite scroll operations (matching Java)
	FindFirstPrayerTitleInfosByRoomID(ctx context.Context, roomID uint64, limit int) ([]*PrayerTitleInfo, error)
	FindPrayerTitleInfosByRoomIDAfterTime(ctx context.Context, roomID uint64, afterTime time.Time, limit int) ([]*PrayerTitleInfo, error)

	// PrayerContent operations
	CreatePrayerContent(ctx context.Context, content *PrayerContent) error
	CreatePrayerContents(ctx context.Context, contents []*PrayerContent) error
	FindPrayerContentByID(ctx context.Context, id uint64) (*PrayerContent, error)
	FindPrayerContentsByTitleID(ctx context.Context, titleID uint64) ([]*PrayerContent, error)
	UpdatePrayerContent(ctx context.Context, content *PrayerContent) error
	DeletePrayerContent(ctx context.Context, id uint64) error
	DeletePrayerContentsByTitleID(ctx context.Context, titleID uint64) error

	// PrayerCompletion operations
	CreatePrayerCompletion(ctx context.Context, completion *PrayerCompletion) error
	FindPrayerCompletionsByTitleID(ctx context.Context, titleID uint64) ([]*PrayerCompletion, error)
	ExistsPrayerCompletionByMemberAndTitle(ctx context.Context, memberID, titleID uint64) (bool, error)
}
