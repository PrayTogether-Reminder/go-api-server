package constants

// MemberConstants defines validation rules and limits for Member entity
const (
	MemberEmailMaxLength    = 255 // 이메일 최대 길이
	MemberNameMaxLength     = 100 // 이름 최대 길이
	MemberPasswordMaxLength = 255 // 비밀번호 해시 최대 길이
	MemberPasswordMinLength = 8   // 비밀번호 최소 길이
)

// Validation messages for Member
const (
	ErrMemberEmailEmpty       = "member email cannot be empty"
	ErrMemberEmailTooLong     = "member email exceeds maximum length"
	ErrMemberEmailInvalid     = "member email format is invalid"
	ErrMemberNameEmpty        = "member name cannot be empty"
	ErrMemberNameTooLong      = "member name exceeds maximum length"
	ErrMemberPasswordEmpty    = "member password cannot be empty"
	ErrMemberPasswordTooShort = "member password is too short"
	ErrMemberPasswordTooLong  = "member password exceeds maximum length"
)

// MemberRoomConstants defines validation rules for MemberRoom entity
const (
	MemberRoomRoleMaxLength = 20 // Role 필드 최대 길이
)

// RoomRole values
const (
	RoomRoleOwner  = "OWNER"  // 방 소유자
	RoomRoleAdmin  = "ADMIN"  // 방 관리자
	RoomRoleMember = "MEMBER" // 일반 멤버
)
