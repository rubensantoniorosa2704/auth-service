package domain

import (
	"regexp"
	"strings"
)

// emailRegex is a basic but robust email validation pattern.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email is a value object that guarantees a structurally valid email address.
type Email struct {
	value string
}

// NewEmail validates and creates an Email value object.
func NewEmail(value string) (Email, error) {
	value = strings.TrimSpace(strings.ToLower(value))

	if !emailRegex.MatchString(value) {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: value}, nil
}

// String returns the email address as a plain string.
func (e Email) String() string {
	return e.value
}
