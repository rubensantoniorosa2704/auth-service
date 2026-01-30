package ports

import "github.com/rubensantoniorosa2704/auth-service/internal/core/domain"

type UserRepository interface {
	Save(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
}
