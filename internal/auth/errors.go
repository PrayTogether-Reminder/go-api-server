package auth

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
)

const (
	incorrectEmailPassword = "INCORRECT_EMAIL_PASSWORD" // errInfo
	refreshTokenNotFound   = "REFRESH_TOKEN_NOT_FOUND"  // errInfo
	refreshTokenExpired    = "REFRESH_TOKEN_EXPIRED"    // errInfo
	refreshTokenMismatch   = "REFRESH_TOKEN_MISMATCH"   // errInfo
)

var (
	ErrInCorrectEmailPassword = sharedError.NewDomainError(incorrectEmailPassword)
	ErrRefreshTokenNotFound   = sharedError.NewDomainError(refreshTokenNotFound)
	ErrRefreshTokenExpired    = sharedError.NewDomainError(refreshTokenExpired)
	ErrRefreshTokenMismatch   = sharedError.NewDomainError(refreshTokenMismatch)
)

func init() {
	sharedError.RegisterDomainErrorResponse(incorrectEmailPassword, sharedError.ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    "AUTH-003",
		Message: "이메일 또는 비밀번호가 일치하지 않습니다.",
	})

	sharedError.RegisterDomainErrorResponse(refreshTokenNotFound, sharedError.ErrorResponse{
		Status:  http.StatusUnauthorized,
		Code:    "AUTH-007",
		Message: "다시 로그인해 주세요.",
	})

	sharedError.RegisterDomainErrorResponse(refreshTokenExpired, sharedError.ErrorResponse{
		Status:  http.StatusUnauthorized,
		Code:    "AUTH-008",
		Message: "다시 로그인해 주세요.",
	})

	sharedError.RegisterDomainErrorResponse(refreshTokenMismatch, sharedError.ErrorResponse{
		Status:  http.StatusUnauthorized,
		Code:    "AUTH-009",
		Message: "다시 로그인해 주세요.",
	})
}
