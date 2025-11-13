package error

import "net/http"

const (
	invalidTimeFormat = "INVALID_TIME_FORMAT" // errInfo
)

var (
	// ErrInvalidTimeFormat is returned when time format is invalid
	ErrInvalidTimeFormat = NewDomainError(invalidTimeFormat)
)

func init() {
	RegisterDomainErrorResponse(invalidTimeFormat, ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    "COMMON-001",
		Message: "잘못된 시간 형식입니다.",
	})
}
