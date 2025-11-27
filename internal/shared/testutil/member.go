package testutil

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	// TestPassword is the plain text password used for test members
	TestPassword = "password123"
)

// CreateTestMember creates a test member with index 0 (default).
// Uses TestPassword as the password (bcrypt hashed).
func CreateTestMember(t *testing.T, db *gorm.DB) *model.Member {
	t.Helper()
	return CreateTestMemberWithIndex(t, db, 0)
}

// CreateTestMemberWithIndex creates a test member with the given index.
// This allows creating multiple unique test members with sequential data.
// Uses TestPassword as the password (bcrypt hashed).
func CreateTestMemberWithIndex(t *testing.T, db *gorm.DB, index int) *model.Member {
	t.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(TestPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	suffix := ""
	if index > 0 {
		suffix = strconv.Itoa(index)
	}

	testMember := &model.Member{
		Name:        "Test_User_" + suffix,
		Email:       "test" + suffix + "@example.com",
		PhoneNumber: fmt.Sprintf("010-1234-%04d", 5678+index),
		Password:    string(hashedPassword),
	}

	if err := db.Create(testMember).Error; err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}

	return testMember
}

// NewMemberRepository creates a new MemberRepository for testing
func NewMemberRepository() *member.MemberRepository {
	return member.NewMemberRepository()
}

// NewMemberService creates a new MemberService for testing
func NewMemberService(memberRepo *member.MemberRepository) *member.MemberService {
	return member.NewMemberService(memberRepo)
}

// NewMemberUseCase creates a new MemberUseCase for testing
func NewMemberUseCase(db *gorm.DB, memberService *member.MemberService) *member.MemberUseCase {
	return member.NewMemberUseCase(db, memberService)
}
