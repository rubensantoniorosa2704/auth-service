package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rubensantoniorosa2704/auth-service/internal/adapters/db/generated"
	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
)

// PostgresUserRepository implements ports.UserRepository using PostgreSQL.
type PostgresUserRepository struct {
	pool *pgxpool.Pool
	q    *generated.Queries
}

// NewPostgresUserRepository creates a new repository backed by a pgxpool.
func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		pool: pool,
		q:    generated.New(pool),
	}
}

// Save persists a domain User into the database.
func (r *PostgresUserRepository) Save(ctx context.Context, user *domain.User) error {
	var uuidPG pgtype.UUID
	if err := uuidPG.Scan(user.ID); err != nil {
		return fmt.Errorf("scanning uuid: %w", err)
	}

	createdAtPG := pgtype.Timestamptz{Time: user.CreatedAt, Valid: true}

	_, err := r.q.CreateUser(ctx, generated.CreateUserParams{
		ID:           uuidPG,
		Email:        user.Email.String(),
		PasswordHash: user.PasswordHash.String(),
		CreatedAt:    createdAtPG,
		UpdatedAt:    createdAtPG,
	})
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	return nil
}

// FindByEmail retrieves a domain User by email address.
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	var idStr string
	if u.ID.Valid {
		parsed, err := uuid.FromBytes(u.ID.Bytes[:])
		if err != nil {
			return nil, fmt.Errorf("parsing user id: %w", err)
		}
		idStr = parsed.String()
	}

	emailVO, err := domain.NewEmail(u.Email)
	if err != nil {
		return nil, fmt.Errorf("reconstructing email: %w", err)
	}

	passVO, err := domain.NewPasswordHash(u.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("reconstructing password hash: %w", err)
	}

	return &domain.User{
		ID:           idStr,
		Email:        emailVO,
		PasswordHash: passVO,
		CreatedAt:    u.CreatedAt.Time,
	}, nil
}
