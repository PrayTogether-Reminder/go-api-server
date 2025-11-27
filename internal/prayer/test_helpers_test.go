package prayer_test

import (
	"testing"

	testutil2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/prayer"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	"gorm.io/gorm"
)

// setupTestEnvironment creates all dependencies needed for prayer handler tests
// Returns the handler, database, and a test member
func setupTestEnvironment(t *testing.T) (*prayer.PrayerHandler, *gorm.DB, int64) {
	t.Helper()

	// Setup test database
	db := testutil2.SetupTestDB(t)
	t.Cleanup(func() {
		testutil2.CleanupTestDB(t, db)
	})

	// Create a test member for authenticated requests
	testMember := testutil2.CreateTestMember(t, db)

	// Setup room dependencies (needed for validation)
	roomRepo := room.NewRoomRepository()
	memberRoomRepo := room.NewMemberRoomRepository()
	memberRepo := testutil2.NewMemberRepository()
	memberService := testutil2.NewMemberService(memberRepo)
	roomService := room.NewRoomService(roomRepo, memberRoomRepo, memberService)

	// Setup prayer dependencies
	prayerRepo := prayer.NewPrayerRepository()
	prayerService := prayer.NewPrayerService(prayerRepo)
	prayerUseCase := prayer.NewPrayerUseCase(db, prayerService, roomService, memberService)
	prayerHandler := prayer.NewPrayerHandler(prayerUseCase)

	return prayerHandler, db, int64(testMember.ID)
}
