package error

import "net/http"

// ErrorResponse is the JSON response structure for errors
type ErrorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorSpec represents a reusable error template used to build consistent responses
type ErrorSpec struct {
	Status  int
	Code    string
	Message string
}

// Response returns an ErrorResponse using the spec's metadata and an optional message override
func (s ErrorSpec) Response(message string) ErrorResponse {
	if message == "" {
		message = s.Message
	}
	return ErrorResponse{
		Status:  s.Status,
		Code:    s.Code,
		Message: message,
	}
}

// Common errors
var (
	// ValidationFailed indicates the request payload failed validation
	ValidationFailed = ErrorSpec{
		Status:  http.StatusBadRequest,
		Code:    "METHOD_ARGUMENT_NOT_VALID",
		Message: "잘못된 요청입니다.",
	}

	// InvalidRequest indicates the request format is invalid (e.g., JSON parsing error)
	InvalidRequest = ErrorSpec{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_REQUEST",
		Message: "잘못된 요청 형식입니다.",
	}

	// InternalServerError indicates an unexpected server error
	InternalServerError = ErrorSpec{
		Status:  http.StatusInternalServerError,
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "서버 내부 오류가 발생했습니다.",
	}
)
