package errors

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// GlobalErrorHandler is a middleware that handles all errors
func GlobalErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			handleError(c, err.Err)
		}
	}
}

// HandleError handles errors and sends appropriate response
func HandleError(c *gin.Context, err error) {
	handleError(c, err)
}

func handleError(c *gin.Context, err error) {
	// Prevent double response
	if c.Writer.Written() {
		return
	}

	// Check if it's a BaseError
	var baseErr BaseError
	if errors.As(err, &baseErr) {
		log.Println(baseErr.LogMessage())
		c.JSON(baseErr.Status(), FromBaseError(baseErr))
		return
	}

	// Check if it's a validation error
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		logValidationError(validationErrs)
		c.JSON(http.StatusBadRequest, FromValidationError(validationErrs))
		return
	}

	// Check if it's a JSON syntax error
	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		log.Printf("[ERROR] JSON 구문 오류: %v (위치: %d)", syntaxError.Error(), syntaxError.Offset)
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			MethodArgumentNotValid.Status,
			MethodArgumentNotValid.Code,
			"올바르지 않은 요청입니다.",
		))
		return
	}

	// Check if it's a JSON type error
	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) {
		log.Printf("[ERROR] 타입 변환 실패: '%s = %v' 은(는) '%s' 타입으로 변환할 수 없습니다.",
			typeError.Field, typeError.Value, typeError.Type)
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			MethodArgumentTypeMismatch.Status,
			MethodArgumentTypeMismatch.Code,
			"올바르지 않은 요청입니다.",
		))
		return
	}

	// Default error handling
	log.Printf("[ERROR] 정의되지 않은 예외 발생: %v", err)
	c.JSON(http.StatusInternalServerError, NewErrorResponse(
		InternalServerError.Status,
		InternalServerError.Code,
		"알 수 없는 오류가 발생했습니다.",
	))
}

// RecoveryMiddleware handles panics and converts them to errors
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] Recovered from panic: %v", r)

				var err error
				switch x := r.(type) {
				case string:
					err = errors.New(x)
				case error:
					err = x
				default:
					err = errors.New("unknown panic")
				}

				handleError(c, err)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ValidationErrorMiddleware handles validation errors from binding
func ValidationErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there was a binding error
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				if e.Type == gin.ErrorTypeBind {
					handleBindingError(c, e.Err)
					return
				}
			}
		}
	}
}

func handleBindingError(c *gin.Context, err error) {
	// Check for validation errors
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		logValidationError(validationErrs)
		c.JSON(http.StatusBadRequest, FromValidationError(validationErrs))
		return
	}

	// Check for JSON unmarshal errors
	if strings.Contains(err.Error(), "cannot unmarshal") {
		log.Printf("[ERROR] 타입 변환 실패: %v", err)
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			MethodArgumentTypeMismatch.Status,
			MethodArgumentTypeMismatch.Code,
			"올바르지 않은 요청입니다.",
		))
		return
	}

	// Check for missing required fields
	if strings.Contains(err.Error(), "required") {
		log.Printf("[ERROR] 필수 필드 누락: %v", err)
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			MethodArgumentNotValid.Status,
			MethodArgumentNotValid.Code,
			"필수 항목이 누락되었습니다.",
		))
		return
	}

	// Default binding error
	log.Printf("[ERROR] 바인딩 오류: %v", err)
	c.JSON(http.StatusBadRequest, NewErrorResponse(
		MethodArgumentNotValid.Status,
		MethodArgumentNotValid.Code,
		"올바르지 않은 요청입니다.",
	))
}

func logValidationError(validationErrs validator.ValidationErrors) {
	var messages []string
	for _, e := range validationErrs {
		messages = append(messages, e.Field()+"="+e.Tag())
	}
	log.Printf("[ERROR] 유효성 검사 실패 : [ %s ]", strings.Join(messages, ", "))
}
