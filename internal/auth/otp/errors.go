package otp

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
)

const otpSendFailed = "OTP_SEND_FAILED"

var ErrSendFailed = sharedError.NewDomainError(otpSendFailed)

func init() {
	sharedError.RegisterDomainErrorResponse(otpSendFailed, sharedError.ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    "OTP-001",
		Message: "인증 번호 발송에 실패했습니다.",
	})
}
