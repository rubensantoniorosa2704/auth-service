package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
	"github.com/rubensantoniorosa2704/auth-service/internal/core/ports"

	"github.com/google/uuid"
)

// AuthServicer defines the application-level contract for authentication.
// Adapters (e.g., gRPC handler) depend on this interface, not the concrete type.
type AuthServicer interface {
	Register(ctx context.Context, email, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
}

// AuthService implements the core authentication use cases.
type AuthService struct {
	users  ports.UserRepository
	hasher ports.PasswordHasher
	tokens ports.TokenService
	logger *slog.Logger
}

// NewAuthService creates a new AuthService with the required dependencies.
func NewAuthService(
	users ports.UserRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenService,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		users:  users,
		hasher: hasher,
		tokens: tokens,
		logger: logger,
	}
}

// Register creates a new user account after validating inputs and checking for duplicates.
func (s *AuthService) Register(ctx context.Context, email, password string) (string, error) {
	emailVO, err := domain.NewEmail(email)
	if err != nil {
		return "", fmt.Errorf("validating email: %w", err)
	}

	if _, err := domain.NewPassword(password); err != nil {
		return "", fmt.Errorf("validating password: %w", err)
	}

	if _, err := s.users.FindByEmail(ctx, emailVO.String()); err == nil {
		return "", domain.ErrUserAlreadyExists
	}

	hashed, err := s.hasher.Hash(ctx, password)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to hash password", slog.String("email", emailVO.String()), slog.String("error", err.Error()))
		return "", fmt.Errorf("hashing password: %w", err)
	}

	passVO, err := domain.NewPasswordHash(hashed)
	if err != nil {
		return "", fmt.Errorf("creating password hash: %w", err)
	}

	userID := uuid.NewString()
	user := &domain.User{
		ID:           userID,
		Email:        emailVO,
		PasswordHash: passVO,
		CreatedAt:    time.Now(),
	}

	if err := s.users.Save(ctx, user); err != nil {
		s.logger.ErrorContext(ctx, "failed to save user", slog.String("email", emailVO.String()), slog.String("error", err.Error()))
		return "", fmt.Errorf("saving user: %w", err)
	}

	s.logger.InfoContext(ctx, "user registered", slog.String("user_id", userID), slog.String("email", emailVO.String()))
	return userID, nil
}

// Login authenticates a user and returns a signed JWT token.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	ok, err := s.hasher.Verify(ctx, password, user.PasswordHash.String())
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to verify password", slog.String("email", email), slog.String("error", err.Error()))
		return "", domain.ErrInvalidCredentials
	}

	if !ok {
		return "", domain.ErrInvalidCredentials
	}

	token, err := s.tokens.Generate(ctx, user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate token", slog.String("user_id", user.ID), slog.String("error", err.Error()))
		return "", fmt.Errorf("generating token: %w", err)
	}

	s.logger.InfoContext(ctx, "user logged in", slog.String("user_id", user.ID))
	return token, nil
}
