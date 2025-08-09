package test

import (
	"testing"
)

// TestAll runs all integration tests
// This is a convenience function to run all test suites at once
func TestAll(t *testing.T) {
	// Example tests
	t.Run("Example", TestExampleIntegration)

	// Member domain tests
	t.Run("MemberProfileFetch", TestMemberProfileFetchIntegration)

	// Room domain tests
	t.Run("RoomCreate", TestRoomCreateIntegration)
	t.Run("RoomDelete", TestRoomDeleteIntegration)
	t.Run("RoomInfiniteScroll", TestRoomInfiniteScrollIntegration)
	t.Run("RoomMemberFetch", TestRoomMemberFetchIntegration)

	// Prayer domain tests
	t.Run("PrayerCreate", TestPrayerCreateIntegration)
	t.Run("PrayerUpdate", TestPrayerUpdateIntegration)
	t.Run("PrayerDelete", TestPrayerDeleteIntegration)
	t.Run("PrayerCompletion", TestPrayerCompletionIntegration)
	t.Run("PrayerContentFetch", TestPrayerContentFetchIntegration)
	t.Run("PrayerInfiniteScroll", TestPrayerInfiniteScrollIntegration)

	// Invitation domain tests
	t.Run("InvitationCreate", TestInvitationCreateIntegration)
	t.Run("InvitationScroll", TestInvitationScrollIntegration)
	t.Run("InvitationUpdateStatus", TestInvitationUpdateStatusIntegration)
}
