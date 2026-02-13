package ports

import "context"

// TokenService defines the contract for token generation and validation.
type TokenService interface {
	Generate(ctx context.Context, userID string) (string, error)
	Validate(ctx context.Context, token string) (string, error)
}
