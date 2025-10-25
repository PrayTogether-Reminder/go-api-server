package member

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

var (
	// 010-1234-5678 또는 01012345678
	phoneRegex = regexp.MustCompile(`^01[0-9]-?[0-9]{4}-?[0-9]{4}$`)
)

func ValidatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	return phoneRegex.MatchString(phone)
}
