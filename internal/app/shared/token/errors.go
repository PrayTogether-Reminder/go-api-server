package token

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/error"
)

const (
	invalidToken         = "INVALID_TOKEN"          // errInfo
	expiredToken         = "EXPIRED_TOKEN"          // errInfo
	invalidClaims        = "INVALID_CLAIMS"         // errInfo
	generateAccessToken  = "GENERATE_ACCESS_TOKEN"  // errInfo
	generateRefreshToken = "GENERATE_REFRESH_TOKEN" // errInfo
)

var (
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = sharedError.NewDomainError(invalidToken)

	// ErrExpiredToken is returned when token is expired
	ErrExpiredToken = sharedError.NewDomainError(expiredToken)

	// ErrInvalidClaims is returned when token claims are invalid
	ErrInvalidClaims = sharedError.NewDomainError(invalidClaims)

	// ErrGenerateAccessToken is returned when access token generation fails
	ErrGenerateAccessToken = sharedError.NewDomainError(generateAccessToken)

	// ErrGenerateRefreshToken is returned when refresh token generation fails
	ErrGenerateRefreshToken = sharedError.NewDomainError(generateRefreshToken)
)

func init() {
	sharedError.RegisterDomainErrorResponse(invalidToken, sharedError.ErrorResponse{
		Status:  http.StatusUnauthorized,
		Code:    "TOKEN-001",
		Message: "로그인을 다시 해주세요.",
	})

	sharedError.RegisterDomainErrorResponse(expiredToken, sharedError.ErrorResponse{
		Status:  http.StatusUnauthorized,
		Code:    "TOKEN-002",
		Message: "로그인을 다시 해주세요.",
	})

	sharedError.RegisterDomainErrorResponse(invalidClaims, sharedError.ErrorResponse{
		Status:  http.StatusUnauthorized,
		Code:    "TOKEN-003",
		Message: "로그인을 다시 해주세요.",
	})

	sharedError.RegisterDomainErrorResponse(generateAccessToken, sharedError.ErrorResponse{
		Status:  http.StatusInternalServerError,
		Code:    "TOKEN-004",
		Message: "인증에 실패했습니다.",
	})

	sharedError.RegisterDomainErrorResponse(generateRefreshToken, sharedError.ErrorResponse{
		Status:  http.StatusInternalServerError,
		Code:    "TOKEN-005",
		Message: "인증에 실패했습니다.",
	})
}
