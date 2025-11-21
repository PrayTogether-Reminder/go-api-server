package testutil

import (
	"testing"

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
	memberID *int64,
	memberName string,
	content string,
) *model.PrayerContent {
	t.Helper()

	prayerContent := &model.PrayerContent{
		PrayerTitleID: prayerTitleID,
		WriterID:      writerID,
		MemberID:      memberID,
		MemberName:    memberName,
		Content:       content,
	}

	if err := db.Create(prayerContent).Error; err != nil {
		t.Fatalf("Failed to create test prayer content: %v", err)
	}

	return prayerContent
}
