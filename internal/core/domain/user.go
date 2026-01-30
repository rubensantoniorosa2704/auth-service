package domain

import "time"

type User struct {
	ID           string
	Email        Email
	PasswordHash PasswordHash
	CreatedAt    time.Time
}
