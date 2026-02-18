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
	users     ports.UserRepository
	hasher    ports.PasswordHasher
	tokens    ports.TokenService
	logger    *slog.Logger
	dummyHash string // Pre-computed hash for timing attack mitigation
}

// NewAuthService creates a new AuthService with the required dependencies.
func NewAuthService(
	users ports.UserRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenService,
	logger *slog.Logger,
) *AuthService {
	// Pre-compute a dummy hash for timing attack mitigation.
	// This hash will be used when a user doesn't exist, ensuring constant-time response.
	dummyHash, err := hasher.Hash(context.Background(), "dummy-password-for-timing-safety")
	if err != nil {
		logger.Warn("failed to generate dummy hash, timing attack mitigation may be affected", slog.String("error", err.Error()))
		dummyHash = "$argon2id$v=19$m=65536,t=1,p=4$c29tZXNhbHQxMjM0NTY$hash" // Fallback
	}

	return &AuthService{
		users:     users,
		hasher:    hasher,
		tokens:    tokens,
		logger:    logger,
		dummyHash: dummyHash,
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
// This implementation is protected against timing attacks by always performing
// password verification, even when the user doesn't exist.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	
	// Determine which hash to verify against.
	// If user doesn't exist, use dummy hash to maintain constant time.
	hashToVerify := s.dummyHash
	var userID string
	userExists := err == nil

	if userExists {
		hashToVerify = user.PasswordHash.String()
		userID = user.ID
	}

	// Always perform password verification, regardless of whether user exists.
	// This prevents timing attacks that could enumerate valid email addresses.
	ok, err := s.hasher.Verify(ctx, password, hashToVerify)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to verify password", slog.String("error", err.Error()))
		return "", domain.ErrInvalidCredentials
	}

	// Only proceed if user exists AND password is correct
	if !userExists || !ok {
		return "", domain.ErrInvalidCredentials
	}

	token, err := s.tokens.Generate(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate token", slog.String("user_id", userID), slog.String("error", err.Error()))
		return "", fmt.Errorf("generating token: %w", err)
	}

	s.logger.InfoContext(ctx, "user logged in", slog.String("user_id", userID))
	return token, nil
}
