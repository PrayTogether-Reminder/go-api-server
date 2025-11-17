package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

// CreateTestRoom creates a single room with member-room relationship for the given member.
// Returns the created room.
func CreateTestRoom(t *testing.T, db *gorm.DB, memberID int64, name, description string) *model.Room {
	t.Helper()

	room := &model.Room{
		Name:        name,
		Description: description,
	}

	if err := db.Create(room).Error; err != nil {
		t.Fatalf("failed to create test room: %v", err)
	}

	memberRoom := &model.MemberRoom{
		MemberID:       memberID,
		RoomID:         room.ID,
		Role:           model.RoomRoleOwner,
		IsNotification: true,
	}

	if err := db.Create(memberRoom).Error; err != nil {
		t.Fatalf("failed to create member_room relationship: %v", err)
	}

	return room
}

// CreateTestRooms creates rooms and member-room relations for pagination-heavy tests.
// Rooms are created with descending timestamps so cursor-based sorting scenarios are easy to verify.
func CreateTestRooms(t *testing.T, db *gorm.DB, memberID int64, count int) {
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

// AddMembersToRoom adds multiple members to a room with MEMBER role.
// Returns the created members.
func AddMembersToRoom(t *testing.T, db *gorm.DB, roomID int64, count int) []*model.Member {
	t.Helper()

	members := make([]*model.Member, 0, count)

	// Get current member count to avoid email conflicts
	var existingCount int64
	db.Model(&model.Member{}).Count(&existingCount)
	baseIndex := int(existingCount) + 1

	for i := 0; i < count; i++ {
		// Create a new member with unique index
		member := CreateTestMemberWithIndex(t, db, baseIndex+i)

		// Add member to room
		memberRoom := &model.MemberRoom{
			MemberID:       member.ID,
			RoomID:         roomID,
			Role:           model.RoomRoleMember,
			IsNotification: true,
		}

		if err := db.Create(memberRoom).Error; err != nil {
			t.Fatalf("failed to create member_room relationship: %v", err)
		}

		members = append(members, member)
	}

	return members
}
