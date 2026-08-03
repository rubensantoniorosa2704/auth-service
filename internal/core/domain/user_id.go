package domain

import "github.com/google/uuid"

// UserID is a value object that guarantees a valid UUID.
type UserID struct {
	value string
}

// NewUserID validates and creates a UserID value object.
// It rejects empty strings and any value that is not a valid UUID.
func NewUserID(raw string) (UserID, error) {
	if raw == "" {
		return UserID{}, ErrInvalidUUID
	}

	if _, err := uuid.Parse(raw); err != nil {
		return UserID{}, ErrInvalidUUID
	}

	return UserID{value: raw}, nil
}

// String returns the UUID as a plain string.
func (u UserID) String() string {
	return u.value
}
