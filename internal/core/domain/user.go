package domain

import "time"

// User is the core domain entity representing an authenticated user.
type User struct {
	ID           string
	Email        Email
	PasswordHash PasswordHash
	CreatedAt    time.Time
	LastLoginAt  *time.Time // Nullable: nil until the user logs in for the first time.
}
