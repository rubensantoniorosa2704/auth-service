package services_test

import (
	"testing"

	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
	"github.com/rubensantoniorosa2704/auth-service/internal/core/services"
)

// =====================
// Test helper
// =====================

func newAuthService() *services.AuthService {
	repo := newFakeUserRepo()
	hasher := &fakeHasher{}
	token := &fakeTokenService{}

	return services.NewAuthService(repo, hasher, token)
}

// =====================
// Tests
// =====================

func TestRegister_Success(t *testing.T) {
	service := newAuthService()

	err := service.Register("test@email.com", "123456")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	service := newAuthService()

	_ = service.Register("test@email.com", "123456")
	err := service.Register("test@email.com", "123456")

	if err != domain.ErrUserAlreadyExists {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	service := newAuthService()

	_ = service.Register("test@email.com", "123456")

	token, err := service.Login("test@email.com", "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected token, got empty string")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	service := newAuthService()

	_ = service.Register("test@email.com", "123456")

	_, err := service.Login("test@email.com", "wrong-password")
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	service := newAuthService()

	_, err := service.Login("notfound@email.com", "123456")
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
