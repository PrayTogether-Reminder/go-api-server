package error

import "net/http"

// ErrorResponse is the JSON response structure for errors
type ErrorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Common errors
var (
	// ValidationFailed indicates the request payload failed validation
	ValidationFailed = ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    "METHOD_ARGUMENT_NOT_VALID",
		Message: "잘못된 요청입니다.",
	}

	// InvalidRequest indicates the request format is invalid (e.g., JSON parsing error)
	InvalidRequest = ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_REQUEST",
		Message: "잘못된 요청 형식입니다.",
	}

	// InternalServerError indicates an unexpected server error
	InternalServerError = ErrorResponse{
		Status:  http.StatusInternalServerError,
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "서버 내부 오류가 발생했습니다.",
	}
)
