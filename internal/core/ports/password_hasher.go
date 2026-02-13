package ports

import "context"

// PasswordHasher defines the contract for password hashing and verification.
type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Verify(ctx context.Context, password string, hash string) (bool, error)
}
