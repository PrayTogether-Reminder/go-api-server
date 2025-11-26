package room_test

import (
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"gorm.io/gorm"
)

// setupTestEnvironment creates all dependencies needed for room handler tests
// Returns the handler, database, and a test member ID
func setupTestEnvironment(t *testing.T) (*room.RoomHandler, *gorm.DB, int64) {
	t.Helper()

	// Setup test database
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestDB(t, db)
	})

	// Create a test member for authenticated requests
	testMember := testutil.CreateTestMember(t, db)

	// Setup dependencies
	roomRepo := room.NewRoomRepository()
	memberRoomRepo := room.NewMemberRoomRepository()

	// Create member service for validation
	memberRepo := testutil.NewMemberRepository()
	memberService := testutil.NewMemberService(memberRepo)

	roomService := room.NewRoomService(roomRepo, memberRoomRepo, memberService)
	roomUseCase := room.NewRoomUseCase(db, roomService)
	roomHandler := room.NewRoomHandler(roomUseCase)

	return roomHandler, db, int64(testMember.ID)
}
