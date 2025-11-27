package invitation

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/error"
)

var (
	// ErrInvitationNotFound is returned when an invitation is not found
	ErrInvitationNotFound = sharedError.NewDomainError("INVITATION_NOT_FOUND")

	// ErrAlreadyRespondedInvitation is returned when trying to respond to an already responded invitation
	ErrAlreadyRespondedInvitation = sharedError.NewDomainError("ALREADY_RESPONDED_INVITATION")

	// ErrInvalidInvitationResponse is returned when an invalid invitation response is provided
	ErrInvalidInvitationResponse = sharedError.NewDomainError("INVALID_INVITATION_RESPONSE")
)

const (
	invitationNotFound         = "INVITATION_NOT_FOUND"
	alreadyRespondedInvitation = "ALREADY_RESPONDED_INVITATION"
	invalidInvitationResponse  = "INVALID_INVITATION_RESPONSE"
)

func init() {
	// Register domain error responses
	sharedError.RegisterDomainErrorResponse(invitationNotFound, sharedError.ErrorResponse{
		Status:  http.StatusNotFound,
		Code:    "INVITATION-001",
		Message: "방 초대장을 찾을 수 없습니다.",
	})

	sharedError.RegisterDomainErrorResponse(alreadyRespondedInvitation, sharedError.ErrorResponse{
		Status:  http.StatusConflict,
		Code:    "INVITATION-002",
		Message: "이미 응답한 초대장입니다.",
	})

	sharedError.RegisterDomainErrorResponse(invalidInvitationResponse, sharedError.ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    "INVITATION-003",
		Message: "유효하지 않은 초대장 응답입니다.",
	})
}
