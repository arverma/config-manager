package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOAuthStateStore struct {
	db *pgxpool.Pool
}

func NewPostgresOAuthStateStore(db *pgxpool.Pool) *PostgresOAuthStateStore {
	return &PostgresOAuthStateStore{db: db}
}

func (s *PostgresOAuthStateStore) CreateOAuthState(ctx context.Context, state, returnTo string, expiresAt time.Time) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth_oauth_states (state, return_to, expires_at)
		VALUES ($1, $2, $3)
	`, state, returnTo, expiresAt)
	if err != nil {
		return fmt.Errorf("store oauth state: %w", err)
	}
	return nil
}

func (s *PostgresOAuthStateStore) ConsumeOAuthState(ctx context.Context, state string) (string, error) {
	var returnTo string
	var expiresAt time.Time
	err := s.db.QueryRow(ctx, `
		SELECT return_to, expires_at
		FROM auth_oauth_states
		WHERE state = $1
	`, state).Scan(&returnTo, &expiresAt)
	if err != nil {
		return "", fmt.Errorf("invalid oauth state")
	}

	_, _ = s.db.Exec(ctx, `DELETE FROM auth_oauth_states WHERE state = $1`, state)

	if time.Now().UTC().After(expiresAt) {
		return "", fmt.Errorf("oauth state expired")
	}
	if returnTo == "" {
		returnTo = "/"
	}
	return returnTo, nil
}
