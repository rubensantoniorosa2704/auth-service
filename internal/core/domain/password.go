package domain

import "errors"

type PasswordHash struct {
	value string
}

func NewPasswordHash(hash string) (PasswordHash, error) {
	if hash == "" {
		return PasswordHash{}, errors.New("password hash cannot be empty")
	}

	return PasswordHash{value: hash}, nil
}

func (p PasswordHash) String() string {
	return p.value
}
