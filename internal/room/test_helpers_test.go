package room_test

import (
	"testing"

	testutil2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	"gorm.io/gorm"
)

// setupTestEnvironment creates all dependencies needed for room handler tests
// Returns the handler, database, and a test member ID
func setupTestEnvironment(t *testing.T) (*room.RoomHandler, *gorm.DB, int64) {
	t.Helper()

	// Setup test database
	db := testutil2.SetupTestDB(t)
	t.Cleanup(func() {
		testutil2.CleanupTestDB(t, db)
	})

	// Create a test member for authenticated requests
	testMember := testutil2.CreateTestMember(t, db)

	// Setup dependencies
	roomRepo := room.NewRoomRepository()
	memberRoomRepo := room.NewMemberRoomRepository()

	// Create member service for validation
	memberRepo := testutil2.NewMemberRepository()
	memberService := testutil2.NewMemberService(memberRepo)

	roomService := room.NewRoomService(roomRepo, memberRoomRepo, memberService)
	roomUseCase := room.NewRoomUseCase(db, roomService)
	roomHandler := room.NewRoomHandler(roomUseCase)

	return roomHandler, db, int64(testMember.ID)
}
