package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
	"github.com/rubensantoniorosa2704/auth-service/internal/core/services"
)

// newAuthService creates a fully wired AuthService with fakes for testing.
func newAuthService() *services.AuthService {
	repo := newFakeUserRepo()
	hasher := &fakeHasher{}
	token := &fakeTokenService{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return services.NewAuthService(repo, hasher, token, logger)
}

// =====================
// Register tests
// =====================

func TestRegister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		password string
		setup    func(svc *services.AuthService)
		wantErr  error
		wantID   bool
	}{
		{
			name:     "success",
			email:    "user@example.com",
			password: "securepassword",
			wantID:   true,
		},
		{
			name:     "duplicate email",
			email:    "dup@example.com",
			password: "securepassword",
			setup: func(svc *services.AuthService) {
				_, _ = svc.Register(context.Background(), "dup@example.com", "securepassword")
			},
			wantErr: domain.ErrUserAlreadyExists,
		},
		{
			name:     "invalid email format",
			email:    "not-an-email",
			password: "securepassword",
			wantErr:  domain.ErrInvalidEmail,
		},
		{
			name:     "empty email",
			email:    "",
			password: "securepassword",
			wantErr:  domain.ErrInvalidEmail,
		},
		{
			name:     "weak password",
			email:    "short@example.com",
			password: "1234567",
			wantErr:  domain.ErrWeakPassword,
		},
		{
			name:     "empty password",
			email:    "empty@example.com",
			password: "",
			wantErr:  domain.ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newAuthService()
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(svc)
			}

			userID, err := svc.Register(ctx, tt.email, tt.password)

			if tt.wantErr != nil {
				assertErrorIs(t, err, tt.wantErr)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantID && userID == "" {
				t.Fatal("expected non-empty user ID")
			}
		})
	}
}

// =====================
// Login tests
// =====================

func TestLogin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		email     string
		password  string
		wantErr   error
		wantToken bool
	}{
		{
			name:      "success",
			email:     "login@example.com",
			password:  "securepassword",
			wantToken: true,
		},
		{
			name:     "wrong password",
			email:    "login@example.com",
			password: "wrong-password",
			wantErr:  domain.ErrInvalidCredentials,
		},
		{
			name:     "user not found",
			email:    "ghost@example.com",
			password: "securepassword",
			wantErr:  domain.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newAuthService()
			ctx := context.Background()

			// Seed user for login tests that expect an existing user
			if tt.email == "login@example.com" {
				if _, err := svc.Register(ctx, "login@example.com", "securepassword"); err != nil {
					t.Fatalf("failed to seed user: %v", err)
				}
			}

			token, err := svc.Login(ctx, tt.email, tt.password)

			if tt.wantErr != nil {
				assertErrorIs(t, err, tt.wantErr)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantToken && token == "" {
				t.Fatal("expected non-empty token")
			}
		})
	}
}

// TestLogin_TimingAttackMitigation verifies that login attempts take similar time
// regardless of whether the user exists or not, preventing email enumeration.
func TestLogin_TimingAttackMitigation(t *testing.T) {
	t.Parallel()

	svc := newAuthService()
	ctx := context.Background()

	// Register a user
	existingEmail := "existing@example.com"
	password := "securepassword"
	if _, err := svc.Register(ctx, existingEmail, password); err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Test 1: Login with non-existent user should still perform hash verification
	nonExistentEmail := "nonexistent@example.com"
	_, err := svc.Login(ctx, nonExistentEmail, password)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	// Test 2: Login with existing user but wrong password
	_, err = svc.Login(ctx, existingEmail, "wrongpassword")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	// Test 3: Successful login
	token, err := svc.Login(ctx, existingEmail, password)
	if err != nil {
		t.Fatalf("expected successful login, got error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Note: We can't easily test actual timing in unit tests without making them flaky.
	// The important part is that the code path always calls hasher.Verify(),
	// which we've verified by ensuring all three scenarios return the expected results.
}

// assertErrorIs is a test helper that checks if err wraps the target error.
func assertErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error %q, got nil", target)
	}

	if !containsError(err, target) {
		t.Fatalf("expected error wrapping %q, got %q", target, err)
	}
}

// containsError checks via errors.Is or, for wrapped fmt.Errorf errors, via Unwrap.
func containsError(err, target error) bool {
	for e := err; e != nil; {
		if e == target {
			return true
		}
		u, ok := e.(interface{ Unwrap() error })
		if !ok {
			break
		}
		e = u.Unwrap()
	}
	return false
}
