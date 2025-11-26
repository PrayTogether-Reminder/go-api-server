package otp

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
)

const (
	otpSendFailed = "OTP_SEND_FAILED"
	otpNotFound   = "OTP_NOT_FOUND"
)

var (
	ErrSendFailed  = sharedError.NewDomainError(otpSendFailed)
	ErrOTPNotFound = sharedError.NewDomainError(otpNotFound)
)

func init() {
	sharedError.RegisterDomainErrorResponse(otpSendFailed, sharedError.ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    "OTP-001",
		Message: "인증 번호 발송에 실패했습니다.",
	})
	sharedError.RegisterDomainErrorResponse(otpNotFound, sharedError.ErrorResponse{
		Status:  http.StatusInternalServerError,
		Code:    "OTP-002",
		Message: "등록되지 않은 OTP 입니다.",
	})
}
