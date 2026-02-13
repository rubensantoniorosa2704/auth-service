package ports

import (
	"context"

	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
)

// UserRepository defines the contract for user persistence.
type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}
