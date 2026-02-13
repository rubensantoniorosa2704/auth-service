package services_test

import (
	"context"
	"errors"

	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
)

// =====================
// Fake UserRepository
// =====================

type fakeUserRepo struct {
	users map[string]*domain.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users: make(map[string]*domain.User),
	}
}

func (f *fakeUserRepo) Save(_ context.Context, user *domain.User) error {
	f.users[user.Email.String()] = user
	return nil
}

func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	user, ok := f.users[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// =====================
// Fake PasswordHasher
// =====================

type fakeHasher struct{}

func (f *fakeHasher) Hash(_ context.Context, password string) (string, error) {
	return "hashed-" + password, nil
}

func (f *fakeHasher) Verify(_ context.Context, password, hash string) (bool, error) {
	return hash == "hashed-"+password, nil
}

// =====================
// Fake TokenService
// =====================

type fakeTokenService struct{}

func (f *fakeTokenService) Generate(_ context.Context, userID string) (string, error) {
	return "token-for-" + userID, nil
}

func (f *fakeTokenService) Validate(_ context.Context, _ string) (string, error) {
	return "", nil
}
