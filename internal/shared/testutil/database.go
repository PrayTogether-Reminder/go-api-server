package testutil

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
// This can be reused across all integration tests
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for tests
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&model.MemberRoom{},
		&model.Room{},
		&model.Member{},
		// Add other models here as needed
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	if err != nil {
		t.Errorf("Failed to get database instance: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		t.Errorf("Failed to close database: %v", err)
	}
}

// TruncateTable truncates a table for test isolation
func TruncateTable(t *testing.T, db *gorm.DB, tableName string) {
	t.Helper()

	if err := db.Exec("DELETE FROM " + tableName).Error; err != nil {
		t.Fatalf("Failed to truncate table %s: %v", tableName, err)
	}
}

const (
	// TestPassword is the plain text password used for test members
	TestPassword = "password123"
)

// CreateTestMember creates a test member with index 0 (default)
// Uses TestPassword as the password (bcrypt hashed)
func CreateTestMember(t *testing.T, db *gorm.DB) *model.Member {
	t.Helper()
	return CreateTestMemberWithIndex(t, db, 0)
}

// CreateTestMemberWithIndex creates a test member with the given index
// This allows creating multiple unique test members with sequential data
// Uses TestPassword as the password (bcrypt hashed)
func CreateTestMemberWithIndex(t *testing.T, db *gorm.DB, index int) *model.Member {
	t.Helper()

	// Hash the test password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(TestPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Generate unique values based on index
	suffix := ""
	if index > 0 {
		suffix = strconv.Itoa(index)
	}

	member := &model.Member{
		Name:        "Test_User_" + suffix,
		Email:       "test" + suffix + "@example.com",
		PhoneNumber: fmt.Sprintf("010-1234-%04d", 5678+index),
		Password:    string(hashedPassword),
	}

	if err := db.Create(member).Error; err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}

	return member
}
