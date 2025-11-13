package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

// CreateTestRooms creates rooms and member-room relations for pagination-heavy tests.
// Rooms are created with descending timestamps so cursor-based sorting scenarios are easy to verify.
func CreateTestRooms(t *testing.T, db *gorm.DB, memberID uint32, count int) {
	t.Helper()

	if count <= 0 {
		return
	}

	baseTime := time.Now().UTC()

	for i := 0; i < count; i++ {
		createdAt := baseTime.Add(-time.Duration(count-i) * time.Second)

		room := &model.Room{
			Name:        fmt.Sprintf("Test Room %d", i+1),
			Description: fmt.Sprintf("Description for room %d", i+1),
		}
		room.CreatedAt = createdAt
		room.UpdatedAt = createdAt

		if err := db.Create(room).Error; err != nil {
			t.Fatalf("failed to create test room: %v", err)
		}

		memberRoom := &model.MemberRoom{
			MemberID:       memberID,
			RoomID:         room.ID,
			Role:           model.RoomRoleOwner,
			IsNotification: true,
		}
		memberRoom.CreatedAt = createdAt
		memberRoom.UpdatedAt = createdAt

		if err := db.Create(memberRoom).Error; err != nil {
			t.Fatalf("failed to create member_room relationship: %v", err)
		}
	}
}
