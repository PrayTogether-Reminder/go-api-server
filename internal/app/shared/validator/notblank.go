package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidateNotBlank checks that a string contains a non-whitespace character.
func ValidateNotBlank(fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() != reflect.String {
		return false
	}

	return strings.TrimSpace(field.String()) != ""
}
