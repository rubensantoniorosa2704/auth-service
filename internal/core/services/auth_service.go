package services

import (
	"time"

	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
	"github.com/rubensantoniorosa2704/auth-service/internal/core/ports"

	"github.com/google/uuid"
)

type AuthService struct {
	users  ports.UserRepository
	hasher ports.PasswordHasher
	tokens ports.TokenService
}

func NewAuthService(
	users ports.UserRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenService,
) *AuthService {
	return &AuthService{
		users:  users,
		hasher: hasher,
		tokens: tokens,
	}
}

func (s *AuthService) Register(email, password string) error {
	_, err := s.users.FindByEmail(email)
	if err == nil {
		return domain.ErrUserAlreadyExists
	}

	hashed, err := s.hasher.Hash(password)
	if err != nil {
		return err
	}

	emailVO, err := domain.NewEmail(email)
	if err != nil {
		return err
	}

	passVO, err := domain.NewPasswordHash(hashed)
	if err != nil {
		return err
	}

	user := &domain.User{
		ID:           uuid.NewString(),
		Email:        emailVO,
		PasswordHash: passVO,
		CreatedAt:    time.Now(),
	}

	return s.users.Save(user)
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.users.FindByEmail(email)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	if !s.hasher.Verify(password, user.PasswordHash.String()) {
		return "", domain.ErrInvalidCredentials
	}

	return s.tokens.Generate(user.ID)
}
