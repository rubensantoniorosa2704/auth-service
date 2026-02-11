package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rubensantoniorosa2704/auth-service/internal/adapters/db/generated"
	"github.com/rubensantoniorosa2704/auth-service/internal/core/domain"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
	q    *generated.Queries
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		pool: pool,
		q:    generated.New(pool),
	}
}

func (r *PostgresUserRepository) Save(user *domain.User) error {
	var uuidPG pgtype.UUID
	if err := uuidPG.Scan(user.ID); err != nil {
		return err
	}

	createdAtPG := pgtype.Timestamptz{Time: user.CreatedAt, Valid: true}

	_, err := r.q.CreateUser(context.Background(), generated.CreateUserParams{
		ID:           uuidPG,
		Email:        user.Email.String(),
		PasswordHash: user.PasswordHash.String(),
		CreatedAt:    createdAtPG,
		UpdatedAt:    createdAtPG,
	})
	return err
}

func (r *PostgresUserRepository) FindByEmail(email string) (*domain.User, error) {
	u, err := r.q.GetUserByEmail(context.Background(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	var idStr string
	if err := u.ID.Scan(&idStr); err != nil {
		return nil, err
	}

	emailVO, _ := domain.NewEmail(u.Email)
	passVO, _ := domain.NewPasswordHash(u.PasswordHash)

	return &domain.User{
		ID:           idStr,
		Email:        emailVO,
		PasswordHash: passVO,
		CreatedAt:    u.CreatedAt.Time,
	}, nil
}
