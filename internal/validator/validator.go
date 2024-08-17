package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	FieldErrors    map[string]string
	NonFieldErrors []string
}

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func (validator *Validator) Valid() bool {
	return len(validator.FieldErrors) == 0 && len(validator.NonFieldErrors) == 0
}

func (validator *Validator) AddFieldError(key, message string) {
	if validator.FieldErrors == nil {
		validator.FieldErrors = make(map[string]string)
	}

	if _, exists := validator.FieldErrors[key]; !exists {
		validator.FieldErrors[key] = message
	}
}

func (validator *Validator) AddNonFieldError(message string) {
	validator.NonFieldErrors = append(validator.NonFieldErrors, message)
}

func (validator *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		validator.AddFieldError(key, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}
