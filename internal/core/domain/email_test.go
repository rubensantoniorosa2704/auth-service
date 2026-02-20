package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
)

func TestNewEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		// Valid
		{
			name:  "simple valid email",
			input: "user@example.com",
			want:  "user@example.com",
		},
		{
			name:  "uppercase is normalized to lowercase",
			input: "User@Example.COM",
			want:  "user@example.com",
		},
		{
			name:  "leading and trailing whitespace is trimmed",
			input: "  user@example.com  ",
			want:  "user@example.com",
		},
		{
			name:  "plus sign in local part",
			input: "user+tag@example.com",
			want:  "user+tag@example.com",
		},
		{
			name:  "dot in local part",
			input: "first.last@example.com",
			want:  "first.last@example.com",
		},
		{
			name:  "subdomain",
			input: "user@mail.example.com",
			want:  "user@mail.example.com",
		},
		{
			name:  "hyphen in domain",
			input: "user@my-domain.com",
			want:  "user@my-domain.com",
		},
		{
			name:  "two-letter TLD",
			input: "user@example.io",
			want:  "user@example.io",
		},

		// Invalid — empty / blank
		{
			name:    "empty string",
			input:   "",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "only whitespace",
			input:   "   ",
			wantErr: domain.ErrInvalidEmail,
		},

		// Invalid — structure
		{
			name:    "missing @",
			input:   "userexample.com",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "multiple @ signs",
			input:   "user@@example.com",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "missing local part",
			input:   "@example.com",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "missing domain",
			input:   "user@",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "missing TLD",
			input:   "user@example",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "single-letter TLD",
			input:   "user@example.c",
			wantErr: domain.ErrInvalidEmail,
		},

		// Invalid — dots
		{
			name:    "consecutive dots in local part",
			input:   "us..er@example.com",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "consecutive dots in domain",
			input:   "user@exam..ple.com",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "local part starts with dot",
			input:   ".user@example.com",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "local part ends with dot",
			input:   "user.@example.com",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "domain starts with dot",
			input:   "user@.example.com",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "domain ends with dot",
			input:   "user@example.com.",
			wantErr: domain.ErrInvalidEmail,
		},

		// Invalid — hyphens in domain
		{
			name:    "domain starts with hyphen",
			input:   "user@-example.com",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "domain ends with hyphen",
			input:   "user@example-.com",
			wantErr: domain.ErrInvalidEmail,
		},

		// Invalid — length limits
		{
			name:    "local part exceeds 64 characters",
			input:   strings.Repeat("a", 65) + "@example.com",
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "total length exceeds 254 characters",
			input:   strings.Repeat("a", 50) + "@" + strings.Repeat("b", 200) + ".com",
			wantErr: domain.ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			email, err := domain.NewEmail(tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, email.String())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, email.String())
		})
	}
}
