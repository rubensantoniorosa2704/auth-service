package domain_test

import (
	"testing"

	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "valid UUID v4",
			input:   "550e8400-e29b-41d4-a716-446655440000",
			wantErr: nil,
		},
		{
			name:    "valid UUID v1",
			input:   "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			wantErr: nil,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: domain.ErrInvalidUUID,
		},
		{
			name:    "too short",
			input:   "550e8400-e29b-41d4",
			wantErr: domain.ErrInvalidUUID,
		},
		{
			name:    "too long",
			input:   "550e8400-e29b-41d4-a716-446655440000-extra",
			wantErr: domain.ErrInvalidUUID,
		},
		{
			name:    "missing hyphens",
			input:   "550e8400e29b41d4a716446655440000",
			wantErr: nil,
		},
		{
			name:    "invalid characters",
			input:   "550e8400-e29b-41d4-a716-44665544000Z",
			wantErr: domain.ErrInvalidUUID,
		},
		{
			name:    "arbitrary string",
			input:   "not-a-uuid",
			wantErr: domain.ErrInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id, err := domain.NewUserID(tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, id.String())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.input, id.String())
		})
	}
}
