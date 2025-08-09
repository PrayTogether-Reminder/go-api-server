package errors

import (
	"github.com/go-playground/validator/v10"
)

// ErrorResponse represents the error response sent to clients
type ErrorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(status int, code, message string) ErrorResponse {
	return ErrorResponse{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

// FromBaseError creates an error response from a BaseError
func FromBaseError(err BaseError) ErrorResponse {
	return ErrorResponse{
		Status:  err.Status(),
		Code:    err.Code(),
		Message: err.ClientMessage(),
	}
}

// FromValidationError creates an error response from validation errors
func FromValidationError(err error) ErrorResponse {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		message := getValidationMessage(validationErrs[0])
		return ErrorResponse{
			Status:  MethodArgumentNotValid.Status,
			Code:    MethodArgumentNotValid.Code,
			Message: message,
		}
	}

	return ErrorResponse{
		Status:  MethodArgumentNotValid.Status,
		Code:    MethodArgumentNotValid.Code,
		Message: "유효하지 않은 요청입니다.",
	}
}

// getValidationMessage returns Korean message for validation errors
func getValidationMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()

	// Custom messages for specific fields
	fieldMessages := map[string]string{
		"Email":      "이메일",
		"Password":   "비밀번호",
		"Name":       "이름",
		"Title":      "제목",
		"Content":    "내용",
		"RoomName":   "방 이름",
		"RoomID":     "방 ID",
		"MemberID":   "회원 ID",
		"MemberName": "회원 이름",
	}

	fieldName := fieldMessages[field]
	if fieldName == "" {
		fieldName = field
	}

	switch tag {
	case "required":
		return fieldName + "은(는) 필수입니다."
	case "email":
		return "올바른 이메일 형식이 아닙니다."
	case "min":
		if fe.Type().String() == "string" {
			return fieldName + "은(는) " + param + "자 이상이어야 합니다."
		}
		return fieldName + "은(는) " + param + " 이상이어야 합니다."
	case "max":
		if fe.Type().String() == "string" {
			return fieldName + "은(는) " + param + "자 이하여야 합니다."
		}
		return fieldName + "은(는) " + param + " 이하여야 합니다."
	case "len":
		return fieldName + "은(는) " + param + "자여야 합니다."
	case "oneof":
		return fieldName + "은(는) " + param + " 중 하나여야 합니다."
	case "alpha":
		return fieldName + "은(는) 알파벳만 포함해야 합니다."
	case "alphanum":
		return fieldName + "은(는) 알파벳과 숫자만 포함해야 합니다."
	case "numeric":
		return fieldName + "은(는) 숫자여야 합니다."
	case "number":
		return fieldName + "은(는) 숫자여야 합니다."
	case "datetime":
		return fieldName + "은(는) 올바른 날짜 형식이 아닙니다."
	case "excludes":
		return fieldName + "은(는) " + param + "을(를) 포함할 수 없습니다."
	case "excludesall":
		return fieldName + "은(는) " + param + " 문자를 포함할 수 없습니다."
	case "excludesrune":
		return fieldName + "은(는) " + param + " 문자를 포함할 수 없습니다."
	case "startswith":
		return fieldName + "은(는) " + param + "(으)로 시작해야 합니다."
	case "endswith":
		return fieldName + "은(는) " + param + "(으)로 끝나야 합니다."
	case "contains":
		return fieldName + "은(는) " + param + "을(를) 포함해야 합니다."
	case "containsany":
		return fieldName + "은(는) " + param + " 중 하나를 포함해야 합니다."
	case "containsrune":
		return fieldName + "은(는) " + param + " 문자를 포함해야 합니다."
	case "url":
		return fieldName + "은(는) 올바른 URL 형식이 아닙니다."
	case "uri":
		return fieldName + "은(는) 올바른 URI 형식이 아닙니다."
	case "base64":
		return fieldName + "은(는) 올바른 Base64 형식이 아닙니다."
	case "isbn":
		return fieldName + "은(는) 올바른 ISBN 형식이 아닙니다."
	case "isbn10":
		return fieldName + "은(는) 올바른 ISBN-10 형식이 아닙니다."
	case "isbn13":
		return fieldName + "은(는) 올바른 ISBN-13 형식이 아닙니다."
	case "uuid":
		return fieldName + "은(는) 올바른 UUID 형식이 아닙니다."
	case "uuid3":
		return fieldName + "은(는) 올바른 UUID v3 형식이 아닙니다."
	case "uuid4":
		return fieldName + "은(는) 올바른 UUID v4 형식이 아닙니다."
	case "uuid5":
		return fieldName + "은(는) 올바른 UUID v5 형식이 아닙니다."
	case "ascii":
		return fieldName + "은(는) ASCII 문자만 포함해야 합니다."
	case "printascii":
		return fieldName + "은(는) 출력 가능한 ASCII 문자만 포함해야 합니다."
	case "multibyte":
		return fieldName + "은(는) 멀티바이트 문자를 포함해야 합니다."
	case "datauri":
		return fieldName + "은(는) 올바른 Data URI 형식이 아닙니다."
	case "latitude":
		return fieldName + "은(는) 올바른 위도 값이 아닙니다."
	case "longitude":
		return fieldName + "은(는) 올바른 경도 값이 아닙니다."
	case "ssn":
		return fieldName + "은(는) 올바른 SSN 형식이 아닙니다."
	case "ipv4":
		return fieldName + "은(는) 올바른 IPv4 주소가 아닙니다."
	case "ipv6":
		return fieldName + "은(는) 올바른 IPv6 주소가 아닙니다."
	case "ip":
		return fieldName + "은(는) 올바른 IP 주소가 아닙니다."
	case "cidr":
		return fieldName + "은(는) 올바른 CIDR 표기법이 아닙니다."
	case "cidrv4":
		return fieldName + "은(는) 올바른 CIDR v4 표기법이 아닙니다."
	case "cidrv6":
		return fieldName + "은(는) 올바른 CIDR v6 표기법이 아닙니다."
	case "tcp_addr":
		return fieldName + "은(는) 올바른 TCP 주소가 아닙니다."
	case "tcp4_addr":
		return fieldName + "은(는) 올바른 TCP v4 주소가 아닙니다."
	case "tcp6_addr":
		return fieldName + "은(는) 올바른 TCP v6 주소가 아닙니다."
	case "udp_addr":
		return fieldName + "은(는) 올바른 UDP 주소가 아닙니다."
	case "udp4_addr":
		return fieldName + "은(는) 올바른 UDP v4 주소가 아닙니다."
	case "udp6_addr":
		return fieldName + "은(는) 올바른 UDP v6 주소가 아닙니다."
	case "ip_addr":
		return fieldName + "은(는) 올바른 IP 주소가 아닙니다."
	case "ip4_addr":
		return fieldName + "은(는) 올바른 IP v4 주소가 아닙니다."
	case "ip6_addr":
		return fieldName + "은(는) 올바른 IP v6 주소가 아닙니다."
	case "unix_addr":
		return fieldName + "은(는) 올바른 Unix 주소가 아닙니다."
	case "mac":
		return fieldName + "은(는) 올바른 MAC 주소가 아닙니다."
	case "iscolor":
		return fieldName + "은(는) 올바른 색상 값이 아닙니다."
	case "json":
		return fieldName + "은(는) 올바른 JSON 형식이 아닙니다."
	case "jwt":
		return fieldName + "은(는) 올바른 JWT 형식이 아닙니다."
	case "hostname":
		return fieldName + "은(는) 올바른 호스트명이 아닙니다."
	case "fqdn":
		return fieldName + "은(는) 올바른 FQDN이 아닙니다."
	case "unique":
		return fieldName + "은(는) 중복될 수 없습니다."
	case "eqfield":
		return fieldName + "은(는) " + param + "과(와) 같아야 합니다."
	case "nefield":
		return fieldName + "은(는) " + param + "과(와) 달라야 합니다."
	case "gtfield":
		return fieldName + "은(는) " + param + "보다 커야 합니다."
	case "gtefield":
		return fieldName + "은(는) " + param + " 이상이어야 합니다."
	case "ltfield":
		return fieldName + "은(는) " + param + "보다 작아야 합니다."
	case "ltefield":
		return fieldName + "은(는) " + param + " 이하여야 합니다."
	case "eqcsfield":
		return fieldName + "은(는) " + param + "과(와) 같아야 합니다."
	case "necsfield":
		return fieldName + "은(는) " + param + "과(와) 달라야 합니다."
	case "gtcsfield":
		return fieldName + "은(는) " + param + "보다 커야 합니다."
	case "gtecsfield":
		return fieldName + "은(는) " + param + " 이상이어야 합니다."
	case "ltcsfield":
		return fieldName + "은(는) " + param + "보다 작아야 합니다."
	case "ltecsfield":
		return fieldName + "은(는) " + param + " 이하여야 합니다."
	default:
		return fieldName + "의 형식이 올바르지 않습니다."
	}
}
