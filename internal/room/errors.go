package room

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
)

const (
	roomNotFound            = "ROOM_NOT_FOUND"            // errInfo
	memberRoomNotFound      = "MEMBER_ROOM_NOT_FOUND"     // errInfo
	memberRoomAlreadyExists = "MEMBER_ROOM_ALREADY_EXIST" // errInfo
	someoneAlreadyExists    = "SOMEONE_ALREADY_EXIST"     // errInfo
	invalidRoomID           = "INVALID_ROOM_ID"           // errInfo
	memberAlreadyInRoom     = "MEMBER_ALREADY_IN_ROOM"    // errInfo
)

var (
	ErrRoomNotFound            = sharedError.NewDomainError(roomNotFound)
	ErrMemberRoomNotFound      = sharedError.NewDomainError(memberRoomNotFound)
	ErrMemberRoomAlreadyExists = sharedError.NewDomainError(memberRoomAlreadyExists)
	ErrSomeoneAlreadyExists    = sharedError.NewDomainError(someoneAlreadyExists)
	ErrInvalidRoomID           = sharedError.NewDomainError(invalidRoomID)
	ErrMemberAlreadyInRoom     = sharedError.NewDomainError(memberAlreadyInRoom)
)

func init() {
	// Register domain error responses
	sharedError.RegisterDomainErrorResponse(roomNotFound, sharedError.ErrorResponse{
		Status:  http.StatusNotFound,
		Code:    "ROOM-001",
		Message: "방을 찾을 수 없습니다.",
	})

	sharedError.RegisterDomainErrorResponse(memberRoomNotFound, sharedError.ErrorResponse{
		Status:  http.StatusNotFound,
		Code:    "ROOM-002",
		Message: "회원이 속한 방을 찾을 수 없습니다.",
	})

	sharedError.RegisterDomainErrorResponse(memberRoomAlreadyExists, sharedError.ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    "ROOM-003",
		Message: "이미 방에 존재하는 회원입니다.",
	})

	sharedError.RegisterDomainErrorResponse(someoneAlreadyExists, sharedError.ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    "ROOM-004",
		Message: "이미 방에 존재하는 회원이 있습니다.",
	})

	sharedError.RegisterDomainErrorResponse(invalidRoomID, sharedError.ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    "ROOM-005",
		Message: "잘 못된 방을 선택하셨습니다.",
	})

	sharedError.RegisterDomainErrorResponse(memberAlreadyInRoom, sharedError.ErrorResponse{
		Status:  http.StatusConflict,
		Code:    "ROOM-006",
		Message: "이미 방에 가입된 회원이 있습니다.",
	})
}
