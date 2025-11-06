package member

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
)

const memberAlreadyExists = "MEMBER_ALREADY_EXISTS" // errInfo

var ErrMemberAlreadyExists = sharedError.NewDomainError(memberAlreadyExists)

func init() {
	sharedError.RegisterDomainErrorResponse(memberAlreadyExists, sharedError.ErrorResponse{
		Status:  http.StatusConflict,
		Code:    "MEMBER-002",
		Message: "이미 가입된 사용자입니다.",
	})
}
