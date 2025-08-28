package model

import (
	"errors"
	"regexp"
	"strings"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/constants"
)

// emailRegex for email validation
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Member represents a user in the system
// Oracle sequence MEMBER_SEQ is used for ID generation
type Member struct {
	// Primary key with Oracle sequence
	ID int64 `gorm:"primaryKey;default:MEMBER_SEQ.NEXTVAL"`

	// Core fields
	Email    string `gorm:"column:email;type:VARCHAR2(255);not null;uniqueIndex:idx_member_email"` // 이메일 (unique)
	Name     string `gorm:"column:name;type:VARCHAR2(100);not null"`                               // 이름
	Password string `gorm:"column:password;type:VARCHAR2(255);not null"`                           // 암호화된 비밀번호

	BaseEntity
}

// TableName specifies the table name for Member
func (*Member) TableName() string {
	return "member"
}

// NewMember creates a new Member instance with validation
// Factory method pattern (Java의 static create 메서드와 동일)
func NewMember(name, email, password string) (*Member, error) {
	// Trim whitespace
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	email = strings.ToLower(email) // 이메일은 소문자로 정규화

	// Validation
	if err := validateMemberFields(name, email, password); err != nil {
		return nil, err
	}

	// Note: password should be hashed before storing (handled in service layer)
	return &Member{
		Name:     name,
		Email:    email,
		Password: password, // This should be hashed password
	}, nil
}

// validateMemberFields validates member creation input
func validateMemberFields(name, email, password string) error {
	// Name validation
	if name == "" {
		return errors.New(constants.ErrMemberNameEmpty)
	}
	if len(name) > constants.MemberNameMaxLength {
		return errors.New(constants.ErrMemberNameTooLong)
	}

	// Email validation
	if email == "" {
		return errors.New(constants.ErrMemberEmailEmpty)
	}
	if len(email) > constants.MemberEmailMaxLength {
		return errors.New(constants.ErrMemberEmailTooLong)
	}
	if !emailRegex.MatchString(email) {
		return errors.New(constants.ErrMemberEmailInvalid)
	}

	// Password validation
	if password == "" {
		return errors.New(constants.ErrMemberPasswordEmpty)
	}
	if len(password) < constants.MemberPasswordMinLength {
		return errors.New(constants.ErrMemberPasswordTooShort)
	}
	if len(password) > constants.MemberPasswordMaxLength {
		return errors.New(constants.ErrMemberPasswordTooLong)
	}

	return nil
}
