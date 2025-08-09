package domain

import (
	"errors"
	"time"
	"unicode/utf8"
)

// BaseEntity contains common fields for all entities in member domain
type BaseEntity struct {
	CreatedAt time.Time  `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`
}

// Member represents a member entity
type Member struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Email    string `gorm:"column:email;uniqueIndex;not null;size:255" json:"email"`
	Name     string `gorm:"column:name;not null;size:100" json:"name"`
	Password string `gorm:"column:password;not null" json:"-"` // Hidden in JSON
	BaseEntity
}

// TableName specifies the table name for Member
func (Member) TableName() string {
	return "member"
}

// NewMember creates a new member
func NewMember(name, email, password string) (*Member, error) {
	member := &Member{
		Name:     name,
		Email:    email,
		Password: password,
	}

	if err := member.Validate(); err != nil {
		return nil, err
	}

	return member, nil
}

// Validate validates member data
func (m *Member) Validate() error {
	if utf8.RuneCountInString(m.Name) < 2 || utf8.RuneCountInString(m.Name) > 30 {
		return ErrInvalidMemberName
	}

	if m.Email == "" || len(m.Email) > 100 {
		return ErrInvalidEmail
	}

	if len(m.Password) < 8 {
		return ErrInvalidPassword
	}

	return nil
}

// UpdateName updates member name
func (m *Member) UpdateName(name string) error {
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 30 {
		return ErrInvalidMemberName
	}
	m.Name = name
	return nil
}

// MemberProfile represents member profile information
type MemberProfile struct {
	ID    uint64 `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ToProfile converts Member to MemberProfile
func (m *Member) ToProfile() *MemberProfile {
	return &MemberProfile{
		ID:    m.ID,
		Email: m.Email,
		Name:  m.Name,
	}
}

// MemberInfo represents basic member information
type MemberInfo struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ToInfo converts Member to MemberInfo
func (m *Member) ToInfo() *MemberInfo {
	return &MemberInfo{
		ID:    m.ID,
		Name:  m.Name,
		Email: m.Email,
	}
}

// Domain errors
var (
	ErrMemberNotFound     = errors.New("member not found")
	ErrMemberAlreadyExist = errors.New("member already exists")
	ErrInvalidMemberName  = errors.New("name must be between 2 and 30 characters")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
	ErrInvalidName        = errors.New("invalid name")
)
