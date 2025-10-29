package error

import "net/http"

// ErrorSpec represents a reusable error template used to build consistent responses.
type ErrorSpec struct {
	Status  int
	Code    string
	Message string
}

// Response returns an ErrorResponse using the spec's metadata and an optional message override.
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

var (
	// ValidationFailed indicates the request payload failed validation.
	ValidationFailed = ErrorSpec{
		Status:  http.StatusBadRequest,
		Code:    "METHOD_ARGUMENT_NOT_VALID",
		Message: "잘못된 요청입니다.",
	}
)
