package errors

import "net/http"

// ErrorSpec defines the specification for an error
type ErrorSpec struct {
	Code          string
	Name          string
	Status        int
	DebugMessage  string
	ClientMessage string
}

// Global error specifications
var (
	// Validation errors
	MethodArgumentNotValid = ErrorSpec{
		Code:          "GLOBAL-001",
		Name:          "METHOD_ARGUMENT_NOT_VALID",
		Status:        http.StatusBadRequest,
		DebugMessage:  "유효성 검사 실패",
		ClientMessage: "유효하지 않은 요청입니다.",
	}

	ConstraintViolate = ErrorSpec{
		Code:          "GLOBAL-002",
		Name:          "CONSTRAINT_VIOLATE",
		Status:        http.StatusBadRequest,
		DebugMessage:  "제약 조건 위반",
		ClientMessage: "제약 조건을 위반하였습니다.",
	}

	MethodArgumentTypeMismatch = ErrorSpec{
		Code:          "GLOBAL-003",
		Name:          "METHOD_ARGUMENT_TYPE_MISMATCH",
		Status:        http.StatusBadRequest,
		DebugMessage:  "타입 변환 실패",
		ClientMessage: "올바르지 않은 요청입니다.",
	}

	// Auth errors
	InvalidCredentials = ErrorSpec{
		Code:          "AUTH-001",
		Name:          "INVALID_CREDENTIALS",
		Status:        http.StatusUnauthorized,
		DebugMessage:  "잘못된 인증 정보",
		ClientMessage: "이메일 또는 비밀번호가 올바르지 않습니다.",
	}

	TokenExpired = ErrorSpec{
		Code:          "AUTH-002",
		Name:          "TOKEN_EXPIRED",
		Status:        http.StatusUnauthorized,
		DebugMessage:  "토큰 만료",
		ClientMessage: "인증이 만료되었습니다. 다시 로그인해주세요.",
	}

	TokenInvalid = ErrorSpec{
		Code:          "AUTH-003",
		Name:          "TOKEN_INVALID",
		Status:        http.StatusUnauthorized,
		DebugMessage:  "유효하지 않은 토큰",
		ClientMessage: "유효하지 않은 인증입니다.",
	}

	RefreshTokenNotFound = ErrorSpec{
		Code:          "AUTH-004",
		Name:          "REFRESH_TOKEN_NOT_FOUND",
		Status:        http.StatusUnauthorized,
		DebugMessage:  "리프레시 토큰을 찾을 수 없음",
		ClientMessage: "인증 정보를 찾을 수 없습니다.",
	}

	// Member errors
	MemberNotFound = ErrorSpec{
		Code:          "MEMBER-001",
		Name:          "MEMBER_NOT_FOUND",
		Status:        http.StatusNotFound,
		DebugMessage:  "회원을 찾을 수 없음",
		ClientMessage: "회원을 찾을 수 없습니다.",
	}

	MemberAlreadyExists = ErrorSpec{
		Code:          "MEMBER-002",
		Name:          "MEMBER_ALREADY_EXISTS",
		Status:        http.StatusConflict,
		DebugMessage:  "이미 존재하는 회원",
		ClientMessage: "이미 가입된 이메일입니다.",
	}

	// Room errors
	RoomNotFound = ErrorSpec{
		Code:          "ROOM-001",
		Name:          "ROOM_NOT_FOUND",
		Status:        http.StatusNotFound,
		DebugMessage:  "방을 찾을 수 없음",
		ClientMessage: "방을 찾을 수 없습니다.",
	}

	MemberNotInRoom = ErrorSpec{
		Code:          "ROOM-002",
		Name:          "MEMBER_NOT_IN_ROOM",
		Status:        http.StatusForbidden,
		DebugMessage:  "방에 속하지 않은 회원",
		ClientMessage: "방에 접근 권한이 없습니다.",
	}

	RoomAlreadyExists = ErrorSpec{
		Code:          "ROOM-003",
		Name:          "ROOM_ALREADY_EXISTS",
		Status:        http.StatusConflict,
		DebugMessage:  "이미 존재하는 방",
		ClientMessage: "이미 존재하는 방입니다.",
	}

	// Prayer errors
	PrayerNotFound = ErrorSpec{
		Code:          "PRAYER-001",
		Name:          "PRAYER_NOT_FOUND",
		Status:        http.StatusNotFound,
		DebugMessage:  "기도를 찾을 수 없음",
		ClientMessage: "기도를 찾을 수 없습니다.",
	}

	PrayerTitleNotFound = ErrorSpec{
		Code:          "PRAYER-002",
		Name:          "PRAYER_TITLE_NOT_FOUND",
		Status:        http.StatusNotFound,
		DebugMessage:  "기도 제목을 찾을 수 없음",
		ClientMessage: "기도 제목을 찾을 수 없습니다.",
	}

	PrayerAlreadyCompleted = ErrorSpec{
		Code:          "PRAYER-003",
		Name:          "PRAYER_ALREADY_COMPLETED",
		Status:        http.StatusConflict,
		DebugMessage:  "이미 완료된 기도",
		ClientMessage: "이미 완료된 기도입니다.",
	}

	// Invitation errors
	InvitationNotFound = ErrorSpec{
		Code:          "INVITATION-001",
		Name:          "INVITATION_NOT_FOUND",
		Status:        http.StatusNotFound,
		DebugMessage:  "초대를 찾을 수 없음",
		ClientMessage: "초대를 찾을 수 없습니다.",
	}

	InvitationExpired = ErrorSpec{
		Code:          "INVITATION-002",
		Name:          "INVITATION_EXPIRED",
		Status:        http.StatusGone,
		DebugMessage:  "만료된 초대",
		ClientMessage: "만료된 초대입니다.",
	}

	AlreadyRespondedInvitation = ErrorSpec{
		Code:          "INVITATION-003",
		Name:          "ALREADY_RESPONDED_INVITATION",
		Status:        http.StatusConflict,
		DebugMessage:  "이미 응답한 초대",
		ClientMessage: "이미 응답한 초대입니다.",
	}

	// General errors
	InternalServerError = ErrorSpec{
		Code:          "GENERAL-001",
		Name:          "INTERNAL_SERVER_ERROR",
		Status:        http.StatusInternalServerError,
		DebugMessage:  "서버 내부 오류",
		ClientMessage: "서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요.",
	}

	NotFound = ErrorSpec{
		Code:          "GENERAL-002",
		Name:          "NOT_FOUND",
		Status:        http.StatusNotFound,
		DebugMessage:  "리소스를 찾을 수 없음",
		ClientMessage: "요청한 리소스를 찾을 수 없습니다.",
	}

	Unauthorized = ErrorSpec{
		Code:          "GENERAL-003",
		Name:          "UNAUTHORIZED",
		Status:        http.StatusUnauthorized,
		DebugMessage:  "인증되지 않음",
		ClientMessage: "인증이 필요합니다.",
	}

	Forbidden = ErrorSpec{
		Code:          "GENERAL-004",
		Name:          "FORBIDDEN",
		Status:        http.StatusForbidden,
		DebugMessage:  "접근 권한 없음",
		ClientMessage: "접근 권한이 없습니다.",
	}
)
