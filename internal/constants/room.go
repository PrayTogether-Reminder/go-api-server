package constants

// RoomConstants defines validation rules and limits for Room entity
const (
	RoomNameMaxLength        = 100  // 방 이름 최대 길이
	RoomDescriptionMaxLength = 500  // 방 설명 최대 길이
	RoomDefaultMaxMembers    = 50   // 기본 최대 멤버 수
	RoomMaxMembersLimit      = 1000 // 최대 멤버 수 제한
)

// Validation messages
const (
	ErrRoomNameTooLong        = "room name exceeds maximum length"
	ErrRoomDescriptionTooLong = "room description exceeds maximum length"
	ErrRoomNameEmpty          = "room name cannot be empty"
	ErrRoomDescriptionEmpty   = "room description cannot be empty"
)
