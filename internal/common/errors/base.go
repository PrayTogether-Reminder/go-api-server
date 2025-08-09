package errors

import (
	"fmt"
	"net/http"
	"strings"
)

// BaseError is the interface that all custom errors must implement
type BaseError interface {
	error
	Code() string
	Status() int
	LogMessage() string
	ClientMessage() string
	Fields() map[string]interface{}
}

// AppError is the base implementation of BaseError
type AppError struct {
	spec   ErrorSpec
	fields map[string]interface{}
}

// NewAppError creates a new AppError
func NewAppError(spec ErrorSpec, fields map[string]interface{}) *AppError {
	return &AppError{
		spec:   spec,
		fields: fields,
	}
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.ClientMessage()
}

// Code returns the error code
func (e *AppError) Code() string {
	return e.spec.Code
}

// Status returns the HTTP status code
func (e *AppError) Status() int {
	return e.spec.Status
}

// LogMessage returns a detailed message for logging
func (e *AppError) LogMessage() string {
	var fieldStrs []string
	for k, v := range e.fields {
		fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", k, v))
	}

	fieldsStr := ""
	if len(fieldStrs) > 0 {
		fieldsStr = fmt.Sprintf(" [ %s ]", strings.Join(fieldStrs, ", "))
	}

	return fmt.Sprintf("[ERROR] %s : %s = %s%s",
		e.spec.Code, e.spec.Name, e.spec.DebugMessage, fieldsStr)
}

// ClientMessage returns the message to be sent to the client
func (e *AppError) ClientMessage() string {
	return e.spec.ClientMessage
}

// Fields returns the error fields
func (e *AppError) Fields() map[string]interface{} {
	return e.fields
}

// Is checks if the error is of a specific type
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.spec.Code == t.spec.Code
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, e := range ve.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return strings.Join(messages, "; ")
}

func (ve ValidationErrors) Code() string {
	return "VALIDATION_ERROR"
}

func (ve ValidationErrors) Status() int {
	return http.StatusBadRequest
}

func (ve ValidationErrors) LogMessage() string {
	return fmt.Sprintf("[ERROR] 유효성 검사 실패 : %s", ve.Error())
}

func (ve ValidationErrors) ClientMessage() string {
	if len(ve.Errors) > 0 {
		return ve.Errors[0].Message
	}
	return "유효하지 않은 요청입니다."
}

func (ve ValidationErrors) Fields() map[string]interface{} {
	fields := make(map[string]interface{})
	for _, e := range ve.Errors {
		fields[e.Field] = e.Message
	}
	return fields
}
