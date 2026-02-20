package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
)

// buildUser is a test helper that constructs a valid User with sensible defaults.
// Override only the fields relevant to the test case.
func buildUser(t *testing.T, id, email, hash string) domain.User {
	t.Helper()

	emailVO, err := domain.NewEmail(email)
	require.NoError(t, err)

	hashVO, err := domain.NewPasswordHash(hash)
	require.NoError(t, err)

	return domain.User{
		ID:           id,
		Email:        emailVO,
		PasswordHash: hashVO,
		CreatedAt:    time.Now().UTC(),
	}
}

func TestUser_Fields(t *testing.T) {
	t.Parallel()

	const (
		id    = "550e8400-e29b-41d4-a716-446655440000"
		email = "user@example.com"
		hash  = "$argon2id$v=19$m=65536,t=3,p=4$somesalt$somehash"
	)

	user := buildUser(t, id, email, hash)

	assert.Equal(t, id, user.ID)
	assert.Equal(t, email, user.Email.String())
	assert.Equal(t, hash, user.PasswordHash.String())
	assert.False(t, user.CreatedAt.IsZero(), "CreatedAt must be set")
}

func TestUser_EmailIsNormalized(t *testing.T) {
	t.Parallel()

	// Email normalization (lowercase, trim) happens inside NewEmail.
	// This test ensures User preserves the normalized value end-to-end.
	user := buildUser(t, "some-id", "  USER@EXAMPLE.COM  ", "some-hash")

	assert.Equal(t, "user@example.com", user.Email.String())
}

func TestUser_ZeroValueIsUnusable(t *testing.T) {
	t.Parallel()

	// A zero-value User should have no meaningful data.
	// This documents the expected behavior when User is not constructed via buildUser.
	var user domain.User

	assert.Empty(t, user.ID)
	assert.Empty(t, user.Email.String())
	assert.Empty(t, user.PasswordHash.String())
	assert.True(t, user.CreatedAt.IsZero())
}
