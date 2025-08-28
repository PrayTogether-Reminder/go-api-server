package model

import (
	"errors"
	"strings"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/constants"
)

// Room represents a prayer room where users can share prayers
// Oracle sequence ROOM_SEQ is used for ID generation
type Room struct {
	ID int64 `gorm:"primaryKey;default:ROOM_SEQ.NEXTVAL"`

	Name        string `gorm:"column:name;type:VARCHAR2(100);not null"`        // 방 이름
	Description string `gorm:"column:description;type:VARCHAR2(500);not null"` // 방 설명

	BaseEntity
}

// TableName specifies the table name for Room
func (*Room) TableName() string {
	return "room"
}

// NewRoom creates a new Room instance with validation
func NewRoom(name, description string) (*Room, error) {
	// Trim whitespace
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	// Validation
	if err := validateRoomFields(name, description); err != nil {
		return nil, err
	}

	return &Room{
		Name:        name,
		Description: description,
	}, nil
}

// validateRoomFields validates room creation input
func validateRoomFields(name, description string) error {
	// Name validation
	if name == "" {
		return errors.New(constants.ErrRoomNameEmpty)
	}
	if len(name) > constants.RoomNameMaxLength {
		return errors.New(constants.ErrRoomNameTooLong)
	}

	// Description validation
	if description == "" {
		return errors.New(constants.ErrRoomDescriptionEmpty)
	}
	if len(description) > constants.RoomDescriptionMaxLength {
		return errors.New(constants.ErrRoomDescriptionTooLong)
	}

	return nil
}
