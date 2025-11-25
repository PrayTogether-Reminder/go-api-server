package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

// CreateTestPrayerTitle creates a prayer title in a given room for testing.
// Returns the created prayer title.
func CreateTestPrayerTitle(t *testing.T, db *gorm.DB, roomID int64, title string) *model.PrayerTitle {
	t.Helper()

	prayerTitle := &model.PrayerTitle{
		RoomID: roomID,
		Title:  title,
	}

	if err := db.Create(prayerTitle).Error; err != nil {
		t.Fatalf("Failed to create test prayer title: %v", err)
	}

	return prayerTitle
}

// CreateTestPrayerContent creates a prayer content for a given prayer title for testing.
// Returns the created prayer content.
func CreateTestPrayerContent(
	t *testing.T,
	db *gorm.DB,
	prayerTitleID int64,
	writerID int64,
	writerName string,
	memberID *int64,
	memberName string,
	content string,
) *model.PrayerContent {
	t.Helper()

	prayerContent := &model.PrayerContent{
		PrayerTitleID: prayerTitleID,
		WriterID:      writerID,
		WriterName:    writerName,
		MemberID:      memberID,
		MemberName:    memberName,
		Content:       content,
	}

	if err := db.Create(prayerContent).Error; err != nil {
		t.Fatalf("Failed to create test prayer content: %v", err)
	}

	return prayerContent
}

// SeedPrayerTitlesWithTime creates multiple prayer titles with specific created_time for testing pagination.
// Returns the created prayer titles ordered by creation time (oldest first).
func SeedPrayerTitlesWithTime(t *testing.T, db *gorm.DB, roomID int64, count int) []*model.PrayerTitle {
	t.Helper()

	baseTime := time.Now().UTC()
	titles := make([]*model.PrayerTitle, 0, count)
	for i := 0; i < count; i++ {
		title := CreateTestPrayerTitle(t, db, roomID, fmt.Sprintf("Prayer Title %d", i+1))
		createdAt := baseTime.Add(time.Duration(i) * time.Minute)
		updates := map[string]interface{}{
			"created_time": createdAt,
			"updated_time": createdAt,
		}
		if err := db.Model(title).Updates(updates).Error; err != nil {
			t.Fatalf("failed to adjust prayer title created_time: %v", err)
		}
		title.CreatedAt = createdAt
		titles = append(titles, title)
	}
	return titles
}
