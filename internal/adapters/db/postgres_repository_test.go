package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adapterdb "github.com/rubensantoniorosa2704/auth-service/internal/adapters/db"
	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
)

// buildTestUser constructs a valid domain.User for use in integration tests.
func buildTestUser(t *testing.T, id, email string) *domain.User {
	t.Helper()

	emailVO, err := domain.NewEmail(email)
	require.NoError(t, err)

	hashVO, err := domain.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$somesalt$somehash")
	require.NoError(t, err)

	return &domain.User{
		ID:           id,
		Email:        emailVO,
		PasswordHash: hashVO,
		CreatedAt:    time.Now().UTC().Truncate(time.Millisecond),
	}
}

// =====================
// Save tests
// =====================

func TestPostgresRepository_Save(t *testing.T) {
	t.Parallel()

	pool := newTestDB(t)
	repo := adapterdb.NewPostgresUserRepository(pool)
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func()
		user    func() *domain.User
		wantErr bool
	}{
		{
			name: "saves user successfully",
			user: func() *domain.User {
				return buildTestUser(t, "550e8400-e29b-41d4-a716-446655440001", "save@example.com")
			},
		},
		{
			name: "fails on duplicate email",
			setup: func() {
				u := buildTestUser(t, "550e8400-e29b-41d4-a716-446655440002", "dup@example.com")
				require.NoError(t, repo.Save(ctx, u))
			},
			user: func() *domain.User {
				return buildTestUser(t, "550e8400-e29b-41d4-a716-446655440003", "dup@example.com")
			},
			wantErr: true,
		},
		{
			name: "fails on duplicate id",
			setup: func() {
				u := buildTestUser(t, "550e8400-e29b-41d4-a716-446655440004", "first@example.com")
				require.NoError(t, repo.Save(ctx, u))
			},
			user: func() *domain.User {
				return buildTestUser(t, "550e8400-e29b-41d4-a716-446655440004", "second@example.com")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := repo.Save(ctx, tt.user())

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

// =====================
// FindByEmail tests
// =====================

func TestPostgresRepository_FindByEmail(t *testing.T) {
	t.Parallel()

	pool := newTestDB(t)
	repo := adapterdb.NewPostgresUserRepository(pool)
	ctx := context.Background()

	seeded := buildTestUser(t, "550e8400-e29b-41d4-a716-446655440000", "find@example.com")
	require.NoError(t, repo.Save(ctx, seeded), "failed to seed user")

	tests := []struct {
		name      string
		email     string
		wantErr   error
		wantEmail string
	}{
		{
			name:      "finds existing user by email",
			email:     "find@example.com",
			wantEmail: "find@example.com",
		},
		{
			name:    "returns ErrUserNotFound for unknown email",
			email:   "ghost@example.com",
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:    "returns ErrUserNotFound for empty string",
			email:   "",
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.FindByEmail(ctx, tt.email)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, user)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, user)
			assert.Equal(t, tt.wantEmail, user.Email.String())
			assert.Equal(t, seeded.ID, user.ID)
			assert.NotEmpty(t, user.PasswordHash.String())
		})
	}
}

// =====================
// Round-trip test
// =====================

func TestPostgresRepository_RoundTrip(t *testing.T) {
	t.Parallel()

	pool := newTestDB(t)
	repo := adapterdb.NewPostgresUserRepository(pool)
	ctx := context.Background()

	original := buildTestUser(t, "550e8400-e29b-41d4-a716-446655440099", "roundtrip@example.com")

	require.NoError(t, repo.Save(ctx, original))

	found, err := repo.FindByEmail(ctx, "roundtrip@example.com")
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, original.ID, found.ID)
	assert.Equal(t, original.Email.String(), found.Email.String())
	assert.Equal(t, original.PasswordHash.String(), found.PasswordHash.String())
	// PostgreSQL does not store nanoseconds, so we truncate to milliseconds before comparing.
	assert.WithinDuration(t, original.CreatedAt, found.CreatedAt, time.Millisecond)
}
