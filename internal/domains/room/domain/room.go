package domain

import (
	"errors"
	"time"
	"unicode/utf8"
)

// BaseEntity contains common fields for all entities in room domain
type BaseEntity struct {
	CreatedAt time.Time  `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`
}

// Room represents a room entity
type Room struct {
	ID                    uint64 `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RoomName              string `gorm:"column:room_name;not null;size:100" json:"roomName"`
	Description           string `gorm:"column:description;size:200" json:"description"` // Added to match Java
	IsPrivate             bool   `gorm:"column:is_private;default:false" json:"isPrivate"`
	IsBlocked             bool   `gorm:"column:is_blocked;default:false" json:"isBlocked"`
	PrayStartTime         string `gorm:"column:pray_start_time;not null" json:"prayStartTime"`
	PrayEndTime           string `gorm:"column:pray_end_time;not null" json:"prayEndTime"`
	NotificationStartTime string `gorm:"column:notification_start_time;not null" json:"notificationStartTime"`
	NotificationEndTime   string `gorm:"column:notification_end_time;not null" json:"notificationEndTime"`
	BaseEntity

	// Associations
	Members []RoomMember `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"members,omitempty"`
}

// TableName specifies the table name for Room
func (Room) TableName() string {
	return "room"
}

// NewRoom creates a new room
func NewRoom(
	roomName string,
	description string,
	isPrivate bool,
	prayStartTime, prayEndTime string,
	notificationStartTime, notificationEndTime string,
) (*Room, error) {
	room := &Room{
		RoomName:              roomName,
		Description:           description,
		IsPrivate:             isPrivate,
		PrayStartTime:         prayStartTime,
		PrayEndTime:           prayEndTime,
		NotificationStartTime: notificationStartTime,
		NotificationEndTime:   notificationEndTime,
	}

	if err := room.Validate(); err != nil {
		return nil, err
	}

	return room, nil
}

// Validate validates room data
func (r *Room) Validate() error {
	if utf8.RuneCountInString(r.RoomName) < 2 || utf8.RuneCountInString(r.RoomName) > 50 {
		return ErrInvalidRoomName
	}

	// Validate time formats
	if err := r.validateTimeFormat(r.PrayStartTime); err != nil {
		return ErrInvalidPrayStartTime
	}

	if err := r.validateTimeFormat(r.PrayEndTime); err != nil {
		return ErrInvalidPrayEndTime
	}

	if err := r.validateTimeFormat(r.NotificationStartTime); err != nil {
		return ErrInvalidNotificationStartTime
	}

	if err := r.validateTimeFormat(r.NotificationEndTime); err != nil {
		return ErrInvalidNotificationEndTime
	}

	// Validate time ranges
	if !r.isValidTimeRange(r.PrayStartTime, r.PrayEndTime) {
		return ErrInvalidPrayTimeRange
	}

	if !r.isValidTimeRange(r.NotificationStartTime, r.NotificationEndTime) {
		return ErrInvalidNotificationTimeRange
	}

	return nil
}

// validateTimeFormat validates HH:mm format
func (r *Room) validateTimeFormat(timeStr string) error {
	_, err := time.Parse("15:04", timeStr)
	return err
}

// isValidTimeRange checks if start time is before end time
func (r *Room) isValidTimeRange(startTime, endTime string) bool {
	start, _ := time.Parse("15:04", startTime)
	end, _ := time.Parse("15:04", endTime)
	return start.Before(end)
}

// Block blocks the room
func (r *Room) Block() {
	r.IsBlocked = true
}

// Unblock unblocks the room
func (r *Room) Unblock() {
	r.IsBlocked = false
}

// SetPrivate sets the room as private
func (r *Room) SetPrivate(isPrivate bool) {
	r.IsPrivate = isPrivate
}

// UpdateName updates the room name
func (r *Room) UpdateName(name string) error {
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 50 {
		return ErrInvalidRoomName
	}
	r.RoomName = name
	return nil
}

// UpdatePrayTime updates pray time
func (r *Room) UpdatePrayTime(startTime, endTime string) error {
	if err := r.validateTimeFormat(startTime); err != nil {
		return ErrInvalidPrayStartTime
	}

	if err := r.validateTimeFormat(endTime); err != nil {
		return ErrInvalidPrayEndTime
	}

	if !r.isValidTimeRange(startTime, endTime) {
		return ErrInvalidPrayTimeRange
	}

	r.PrayStartTime = startTime
	r.PrayEndTime = endTime
	return nil
}

// UpdateNotificationTime updates notification time
func (r *Room) UpdateNotificationTime(startTime, endTime string) error {
	if err := r.validateTimeFormat(startTime); err != nil {
		return ErrInvalidNotificationStartTime
	}

	if err := r.validateTimeFormat(endTime); err != nil {
		return ErrInvalidNotificationEndTime
	}

	if !r.isValidTimeRange(startTime, endTime) {
		return ErrInvalidNotificationTimeRange
	}

	r.NotificationStartTime = startTime
	r.NotificationEndTime = endTime
	return nil
}

// AddMember adds a member to the room
func (r *Room) AddMember(memberID uint64, role RoomRole) (*RoomMember, error) {
	// Check if member already exists
	for _, member := range r.Members {
		if member.MemberID == memberID {
			return nil, ErrMemberAlreadyInRoom
		}
	}

	roomMember := &RoomMember{
		RoomID:   r.ID,
		MemberID: memberID,
		Role:     role,
	}

	r.Members = append(r.Members, *roomMember)
	return roomMember, nil
}

// RemoveMember removes a member from the room
func (r *Room) RemoveMember(memberID uint64) error {
	for i, member := range r.Members {
		if member.MemberID == memberID {
			r.Members = append(r.Members[:i], r.Members[i+1:]...)
			return nil
		}
	}
	return ErrMemberNotInRoom
}

// GetMember gets a member from the room
func (r *Room) GetMember(memberID uint64) (*RoomMember, error) {
	for _, member := range r.Members {
		if member.MemberID == memberID {
			return &member, nil
		}
	}
	return nil, ErrMemberNotInRoom
}

// IsOwner checks if a member is the owner of the room
func (r *Room) IsOwner(memberID uint64) bool {
	member, err := r.GetMember(memberID)
	if err != nil {
		return false
	}
	return member.Role == RoleOwner
}

// RoomInfo represents room information
type RoomInfo struct {
	ID                    uint64    `json:"id"`
	RoomName              string    `json:"roomName"`
	IsPrivate             bool      `json:"isPrivate"`
	IsBlocked             bool      `json:"isBlocked"`
	PrayStartTime         string    `json:"prayStartTime"`
	PrayEndTime           string    `json:"prayEndTime"`
	NotificationStartTime string    `json:"notificationStartTime"`
	NotificationEndTime   string    `json:"notificationEndTime"`
	MemberCount           int       `json:"memberCount"`
	CreatedAt             time.Time `json:"createdAt"`
}

// ToInfo converts Room to RoomInfo
func (r *Room) ToInfo() *RoomInfo {
	return &RoomInfo{
		ID:                    r.ID,
		RoomName:              r.RoomName,
		IsPrivate:             r.IsPrivate,
		IsBlocked:             r.IsBlocked,
		PrayStartTime:         r.PrayStartTime,
		PrayEndTime:           r.PrayEndTime,
		NotificationStartTime: r.NotificationStartTime,
		NotificationEndTime:   r.NotificationEndTime,
		MemberCount:           len(r.Members),
		CreatedAt:             r.CreatedAt,
	}
}

// Domain errors
var (
	ErrRoomNotFound                 = errors.New("room not found")
	ErrRoomBlocked                  = errors.New("room is blocked")
	ErrInvalidRoomName              = errors.New("room name must be between 2 and 50 characters")
	ErrInvalidPrayStartTime         = errors.New("invalid pray start time format")
	ErrInvalidPrayEndTime           = errors.New("invalid pray end time format")
	ErrInvalidNotificationStartTime = errors.New("invalid notification start time format")
	ErrInvalidNotificationEndTime   = errors.New("invalid notification end time format")
	ErrInvalidPrayTimeRange         = errors.New("pray start time must be before end time")
	ErrInvalidNotificationTimeRange = errors.New("notification start time must be before end time")
	ErrMemberAlreadyInRoom          = errors.New("member already in room")
	ErrMemberNotInRoom              = errors.New("member not in room")
	ErrNotRoomOwner                 = errors.New("not room owner")
)
