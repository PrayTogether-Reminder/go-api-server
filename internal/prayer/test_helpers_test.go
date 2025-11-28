package prayer_test

import (
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/notification"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/prayer"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	"gorm.io/gorm"
)

// setupTestEnvironment creates all dependencies needed for prayer handler tests
// Returns the handler, database, and a test member
func setupTestEnvironment(t *testing.T) (*prayer.PrayerHandler, *gorm.DB, int64) {
	t.Helper()

	// Setup test database
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestDB(t, db)
	})

	// Create a test member for authenticated requests
	testMember := testutil.CreateTestMember(t, db)

	// Setup room dependencies (needed for validation)
	roomRepo := room.NewRoomRepository()
	memberRoomRepo := room.NewMemberRoomRepository()
	memberRepo := testutil.NewMemberRepository()
	memberService := testutil.NewMemberService(memberRepo)
	roomService := room.NewRoomService(roomRepo, memberRoomRepo, memberService)

	// Setup prayer dependencies
	prayerRepo := prayer.NewPrayerRepository()
	prayerService := prayer.NewPrayerService(prayerRepo)

	// Setup notification dependencies
	notificationRepo := notification.NewNotificationRepository()
	notificationService := notification.NewNotificationService(notificationRepo)

	// Setup FCM Gateway (disabled for testing)
	fcmGateway := &notification.FCMGateway{} // nil client, disabled for tests

	prayerUseCase := prayer.NewPrayerUseCase(db, prayerService, roomService, memberService, notificationService, fcmGateway)
	prayerHandler := prayer.NewPrayerHandler(prayerUseCase)

	return prayerHandler, db, int64(testMember.ID)
}
