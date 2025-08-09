package test

import (
	"fmt"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	memberdomain "pray-together/internal/domains/member/domain"
	prayerdomain "pray-together/internal/domains/prayer/domain"
	roomdomain "pray-together/internal/domains/room/domain"
	"pray-together/internal/infrastructure/jwt"
)

// TestUtils provides utility functions for tests (matching Java TestUtils)
type TestUtils struct {
	db            *gorm.DB
	jwtService    *jwt.JWTService
	emailCounter  int64
	roomCounter   int64
	prayerCounter int64
}

// NewTestUtils creates a new TestUtils instance
func NewTestUtils(db *gorm.DB) *TestUtils {
	// JWT service with test durations
	jwtService := jwt.NewJWTService("test-secret-key", time.Hour, 24*time.Hour)

	return &TestUtils{
		db:         db,
		jwtService: jwtService,
	}
}

// CreateUniqueMember creates a unique member for testing (matching Java createUniqueMember)
func (tu *TestUtils) CreateUniqueMember() *memberdomain.Member {
	counter := atomic.AddInt64(&tu.emailCounter, 1)

	member := &memberdomain.Member{
		Name:     fmt.Sprintf("test%d", counter),
		Email:    fmt.Sprintf("test%d@test.com", counter), // Fixed email format
		Password: "test",                                  // This should be hashed in real implementation
	}

	// Save to database
	result := tu.db.Create(member)
	if result.Error != nil {
		panic(fmt.Sprintf("Failed to create test member: %v", result.Error))
	}

	return member
}

// CreateUniqueRoom creates a unique room for testing (matching Java createUniqueRoom)
func (tu *TestUtils) CreateUniqueRoom() *roomdomain.Room {
	counter := atomic.AddInt64(&tu.roomCounter, 1)

	room, err := roomdomain.NewRoom(
		fmt.Sprintf("test-Room%d", counter),
		fmt.Sprintf("test-Room%d description", counter), // description
		false,            // isPrivate
		"09:00", "21:00", // pray times
		"08:00", "22:00", // notification times
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create room instance: %v", err))
	}

	// Save to database
	result := tu.db.Create(room)
	if result.Error != nil {
		panic(fmt.Sprintf("Failed to create test room: %v", result.Error))
	}

	return room
}

// CreateUniquePrayerTitleWithRoom creates a unique prayer title with room (matching Java createUniquePrayerTitle_With_Room)
func (tu *TestUtils) CreateUniquePrayerTitleWithRoom(room *roomdomain.Room) *prayerdomain.PrayerTitle {
	counter := atomic.AddInt64(&tu.prayerCounter, 1)

	prayerTitle := prayerdomain.NewPrayerTitle(
		room.ID,
		1, // Default creator ID
		fmt.Sprintf("test-prayer-title%d", counter),
	)

	// Save to database
	result := tu.db.Create(prayerTitle)
	if result.Error != nil {
		panic(fmt.Sprintf("Failed to create test prayer title: %v", result.Error))
	}

	return prayerTitle
}

// CreateUniqueMemberRoomWithMemberAndRoom creates a unique member room relationship (matching Java createUniqueMemberRoom_With_Member_AND_Room)
func (tu *TestUtils) CreateUniqueMemberRoomWithMemberAndRoom(member *memberdomain.Member, room *roomdomain.Room) *roomdomain.RoomMember {
	memberRoom := roomdomain.NewRoomMember(room.ID, member.ID, roomdomain.RoleOwner) // Default to OWNER like Java

	// Save to database
	result := tu.db.Create(memberRoom)
	if result.Error != nil {
		panic(fmt.Sprintf("Failed to create test member room: %v", result.Error))
	}

	return memberRoom
}

// CreateAuthHeaderWithMember creates auth header with JWT token for member (matching Java create_Auth_HttpHeader_With_Member)
func (tu *TestUtils) CreateAuthHeaderWithMember(member *memberdomain.Member) map[string]string {
	// Create JWT token
	token, err := tu.jwtService.GenerateAccessToken(member.ID, member.Email, member.Name)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate JWT token: %v", err))
	}

	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json",
	}
}

// GetDB returns the database instance
func (tu *TestUtils) GetDB() *gorm.DB {
	return tu.db
}

// AssertTableCount asserts the count of records in a table
func (tu *TestUtils) AssertTableCount(tableName string, expectedCount int64) {
	var count int64
	tu.db.Table(tableName).Count(&count)

	if count != expectedCount {
		panic(fmt.Sprintf("Expected %d records in table %s, but found %d", expectedCount, tableName, count))
	}
}

// AssertMemberCount asserts the count of members
func (tu *TestUtils) AssertMemberCount(expectedCount int64) {
	tu.AssertTableCount("member", expectedCount)
}

// AssertRoomCount asserts the count of rooms
func (tu *TestUtils) AssertRoomCount(expectedCount int64) {
	tu.AssertTableCount("room", expectedCount)
}

// AssertPrayerTitleCount asserts the count of prayer titles
func (tu *TestUtils) AssertPrayerTitleCount(expectedCount int64) {
	tu.AssertTableCount("prayer_title", expectedCount)
}

// AssertPrayerContentCount asserts the count of prayer contents
func (tu *TestUtils) AssertPrayerContentCount(expectedCount int64) {
	tu.AssertTableCount("prayer_content", expectedCount)
}

// FindAllPrayerTitles returns all prayer titles
func (tu *TestUtils) FindAllPrayerTitles() []prayerdomain.PrayerTitle {
	var titles []prayerdomain.PrayerTitle
	tu.db.Find(&titles)
	return titles
}

// FindAllPrayerContents returns all prayer contents
func (tu *TestUtils) FindAllPrayerContents() []prayerdomain.PrayerContent {
	var contents []prayerdomain.PrayerContent
	tu.db.Find(&contents)
	return contents
}
