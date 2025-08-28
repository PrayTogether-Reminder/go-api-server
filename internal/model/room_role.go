package model

// RoomRole represents the role of a member in a room
type RoomRole string

// RoomRole enum values
const (
	RoomRoleOwner  RoomRole = "OWNER"  // 방 소유자
	RoomRoleMember RoomRole = "MEMBER" // 일반 멤버
)
