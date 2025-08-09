package errors

import (
	"fmt"
	"net/http"
)

// AppError represents application-specific errors
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Details    map[string]interface{}
	Err        error
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap implements errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Err
}

// Common error codes
const (
	// General errors
	CodeInternal     = "INTERNAL_ERROR"
	CodeValidation   = "VALIDATION_ERROR"
	CodeNotFound     = "NOT_FOUND"
	CodeConflict     = "CONFLICT"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeBadRequest   = "BAD_REQUEST"

	// Auth errors
	CodeInvalidCredentials   = "INVALID_CREDENTIALS"
	CodeTokenExpired         = "TOKEN_EXPIRED"
	CodeTokenInvalid         = "TOKEN_INVALID"
	CodeOTPInvalid           = "OTP_INVALID"
	CodeOTPExpired           = "OTP_EXPIRED"
	CodeRefreshTokenNotFound = "REFRESH_TOKEN_NOT_FOUND"

	// Member errors
	CodeMemberNotFound = "MEMBER_NOT_FOUND"
	CodeMemberExists   = "MEMBER_EXISTS"
	CodeEmailExists    = "EMAIL_EXISTS"

	// Room errors
	CodeRoomNotFound     = "ROOM_NOT_FOUND"
	CodeRoomAccessDenied = "ROOM_ACCESS_DENIED"

	// MemberRoom errors
	CodeMemberRoomExists   = "MEMBER_ROOM_EXISTS"
	CodeMemberRoomNotFound = "MEMBER_ROOM_NOT_FOUND"

	// Prayer errors
	CodePrayerNotFound = "PRAYER_NOT_FOUND"

	// Invitation errors
	CodeInvitationNotFound = "INVITATION_NOT_FOUND"
	CodeInvitationExpired  = "INVITATION_EXPIRED"
	CodeAlreadyResponded   = "ALREADY_RESPONDED"

	// FCM errors
	CodeFCMTokenInvalid = "FCM_TOKEN_INVALID"
)

// Constructor functions for common errors

// NewInternalError creates an internal server error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       CodeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, details map[string]interface{}) *AppError {
	return &AppError{
		Code:       CodeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Details:    details,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       CodeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       CodeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       CodeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       CodeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// Auth specific errors

// NewInvalidCredentialsError creates an invalid credentials error
func NewInvalidCredentialsError() *AppError {
	return &AppError{
		Code:       CodeInvalidCredentials,
		Message:    "Invalid email or password",
		StatusCode: http.StatusUnauthorized,
	}
}

// NewTokenExpiredError creates a token expired error
func NewTokenExpiredError() *AppError {
	return &AppError{
		Code:       CodeTokenExpired,
		Message:    "Token has expired",
		StatusCode: http.StatusUnauthorized,
	}
}

// NewTokenInvalidError creates a token invalid error
func NewTokenInvalidError() *AppError {
	return &AppError{
		Code:       CodeTokenInvalid,
		Message:    "Invalid token",
		StatusCode: http.StatusUnauthorized,
	}
}

// Member specific errors

// NewMemberNotFoundError creates a member not found error
func NewMemberNotFoundError() *AppError {
	return &AppError{
		Code:       CodeMemberNotFound,
		Message:    "Member not found",
		StatusCode: http.StatusNotFound,
	}
}

// NewEmailExistsError creates an email exists error
func NewEmailExistsError(email string) *AppError {
	return &AppError{
		Code:       CodeEmailExists,
		Message:    fmt.Sprintf("Email %s already exists", email),
		StatusCode: http.StatusConflict,
	}
}
