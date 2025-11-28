package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidateNotBlank ensures a string contains non-whitespace characters.
func ValidateNotBlank(fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() != reflect.String {
		return false
	}
	return strings.TrimSpace(field.String()) != ""
}
