package tokens

import (
	"context"
	"strings"
	"testing"
)

func TestNewJWTService_SecretValidation(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		issuer    string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid secret with exactly 32 bytes",
			secret:    "12345678901234567890123456789012", // exatamente 32 caracteres
			issuer:    "test-service",
			wantError: false,
		},
		{
			name:      "valid secret with more than 32 bytes",
			secret:    "this-is-a-very-long-secret-with-more-than-32-bytes-for-extra-security",
			issuer:    "test-service",
			wantError: false,
		},
		{
			name:      "empty secret",
			secret:    "",
			issuer:    "test-service",
			wantError: true,
			errorMsg:  "secret cannot be empty",
		},
		{
			name:      "secret too short - 10 bytes",
			secret:    "short-key",
			issuer:    "test-service",
			wantError: true,
			errorMsg:  "secret must be at least 32 bytes long",
		},
		{
			name:      "secret too short - 31 bytes",
			secret:    "almost-valid-but-one-byte-shor",
			issuer:    "test-service",
			wantError: true,
			errorMsg:  "secret must be at least 32 bytes long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewJWTService(tt.secret, tt.issuer)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewJWTService() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("NewJWTService() error = %v, want error containing %q", err, tt.errorMsg)
				}
				if svc != nil {
					t.Errorf("NewJWTService() expected nil service on error, got %v", svc)
				}
			} else {
				if err != nil {
					t.Errorf("NewJWTService() unexpected error = %v", err)
					return
				}
				if svc == nil {
					t.Errorf("NewJWTService() expected service, got nil")
				}
			}
		})
	}
}

func TestJWTService_GenerateAndValidate(t *testing.T) {
	secret := "12345678901234567890123456789012" // exatamente 32 caracteres
	issuer := "test-service"

	svc, err := NewJWTService(secret, issuer)
	if err != nil {
		t.Fatalf("NewJWTService() failed: %v", err)
	}

	ctx := context.Background()
	userID := "test-user-123"

	// Generate token
	token, err := svc.Generate(ctx, userID)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if token == "" {
		t.Error("Generate() returned empty token")
	}

	// Validate token
	extractedUserID, err := svc.Validate(ctx, token)
	if err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	if extractedUserID != userID {
		t.Errorf("Validate() userID = %v, want %v", extractedUserID, userID)
	}
}

func TestJWTService_ValidateInvalidToken(t *testing.T) {
	secret := "12345678901234567890123456789012" // exatamente 32 caracteres
	issuer := "test-service"

	svc, err := NewJWTService(secret, issuer)
	if err != nil {
		t.Fatalf("NewJWTService() failed: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "malformed token",
			token: "not.a.valid.jwt",
		},
		{
			name:  "random string",
			token: "random-string-that-is-not-a-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Validate(ctx, tt.token)
			if err == nil {
				t.Errorf("Validate() expected error for invalid token, got none")
			}
		})
	}
}
