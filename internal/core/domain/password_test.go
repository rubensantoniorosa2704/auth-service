package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
)

func TestNewPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:  "valid password with exactly 8 characters",
			input: "12345678",
		},
		{
			name:  "valid password longer than 8 characters",
			input: "a-very-secure-passphrase",
		},
		{
			name:  "valid password with special characters",
			input: "P@ssw0rd!#$%",
		},
		{
			name:    "empty password",
			input:   "",
			wantErr: domain.ErrWeakPassword,
		},
		{
			name:    "password with 7 characters — one below minimum",
			input:   "1234567",
			wantErr: domain.ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pw, err := domain.NewPassword(tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, pw.String())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.input, pw.String())
		})
	}
}

func TestNewPasswordHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:  "valid argon2 hash",
			input: "$argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaasfasf",
		},
		{
			name:  "any non-empty string is accepted",
			input: "some-hash-value",
		},
		{
			name:    "empty hash",
			input:   "",
			wantErr: domain.ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := domain.NewPasswordHash(tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, hash.String())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.input, hash.String())
		})
	}
}
