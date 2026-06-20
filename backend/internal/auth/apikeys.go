package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type APIKeyStore struct {
	db *pgxpool.Pool
}

func NewAPIKeyStore(db *pgxpool.Pool) *APIKeyStore {
	return &APIKeyStore{db: db}
}

func (s *APIKeyStore) ValidateKey(ctx context.Context, rawKey string) (Actor, error) {
	rawKey = strings.TrimSpace(rawKey)
	if !strings.HasPrefix(rawKey, apiKeyPrefix) {
		return Actor{}, fmt.Errorf("invalid api key format")
	}

	keyHash := hashAPIKey(rawKey)
	var name string
	var revokedAt *time.Time
	err := s.db.QueryRow(ctx, `
		SELECT name, revoked_at
		FROM api_keys
		WHERE key_hash = $1
	`, keyHash).Scan(&name, &revokedAt)
	if err != nil {
		return Actor{}, fmt.Errorf("api key not found")
	}
	if revokedAt != nil {
		return Actor{}, fmt.Errorf("api key revoked")
	}

	_, _ = s.db.Exec(ctx, `UPDATE api_keys SET last_used_at = $2 WHERE key_hash = $1`, keyHash, time.Now().UTC())

	return Actor{Type: ActorTypeAPIKey, ID: name}, nil
}

func (s *APIKeyStore) CreateKey(ctx context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	secret, err := randomSecret(32)
	if err != nil {
		return "", err
	}
	rawKey := apiKeyPrefix + secret
	keyHash := hashAPIKey(rawKey)
	prefix := rawKey
	if len(prefix) > 16 {
		prefix = prefix[:16]
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO api_keys (name, key_hash, key_prefix)
		VALUES ($1, $2, $3)
	`, name, keyHash, prefix)
	if err != nil {
		return "", fmt.Errorf("insert api key: %w", err)
	}

	return rawKey, nil
}

func (s *APIKeyStore) BootstrapFromEnv(ctx context.Context) error {
	raw := strings.TrimSpace(os.Getenv("AUTH_API_KEYS"))
	if raw == "" {
		return nil
	}

	pairs := strings.Split(raw, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid AUTH_API_KEYS entry %q: expected name:key", pair)
		}
		name := strings.TrimSpace(parts[0])
		key := strings.TrimSpace(parts[1])
		if name == "" || key == "" {
			return fmt.Errorf("invalid AUTH_API_KEYS entry %q", pair)
		}
		if err := s.upsertKey(ctx, name, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *APIKeyStore) upsertKey(ctx context.Context, name, rawKey string) error {
	if !strings.HasPrefix(rawKey, apiKeyPrefix) {
		return fmt.Errorf("api key for %q must start with %s", name, apiKeyPrefix)
	}
	keyHash := hashAPIKey(rawKey)
	prefix := rawKey
	if len(prefix) > 16 {
		prefix = prefix[:16]
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO api_keys (name, key_hash, key_prefix)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE
		SET key_hash = EXCLUDED.key_hash,
		    key_prefix = EXCLUDED.key_prefix,
		    revoked_at = NULL
	`, name, keyHash, prefix)
	if err != nil {
		return fmt.Errorf("upsert api key %q: %w", name, err)
	}
	return nil
}

func hashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

func randomSecret(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
