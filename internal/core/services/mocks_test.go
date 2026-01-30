package services_test

import (
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

func (f *fakeUserRepo) Save(user *domain.User) error {
	f.users[user.Email.String()] = user
	return nil
}

func (f *fakeUserRepo) FindByEmail(email string) (*domain.User, error) {
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

func (f *fakeHasher) Hash(password string) (string, error) {
	return "hashed-" + password, nil
}

func (f *fakeHasher) Verify(password, hash string) bool {
	return hash == "hashed-"+password
}

// =====================
// Fake TokenService
// =====================

type fakeTokenService struct{}

func (f *fakeTokenService) Generate(userID string) (string, error) {
	return "token-for-" + userID, nil
}

func (f *fakeTokenService) Validate(token string) (string, error) {
	return "", nil
}
