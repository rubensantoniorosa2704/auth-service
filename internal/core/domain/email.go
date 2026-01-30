package domain

import (
	"errors"
	"strings"
)

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	value = strings.TrimSpace(strings.ToLower(value))

	if !strings.Contains(value, "@") {
		return Email{}, errors.New("invalid email")
	}

	return Email{value: value}, nil
}

func (e Email) String() string {
	return e.value
}
