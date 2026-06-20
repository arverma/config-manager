package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionStore struct {
	db  *pgxpool.Pool
	cfg Config
}

func NewSessionStore(db *pgxpool.Pool, cfg Config) *SessionStore {
	return &SessionStore{db: db, cfg: cfg}
}

type Session struct {
	Token string
	Actor Actor
}

func (s *SessionStore) CreateUserSession(ctx context.Context, email string) (Session, error) {
	token, tokenHash, err := newSessionToken()
	if err != nil {
		return Session{}, err
	}

	now := time.Now().UTC()
	absoluteExpiry := now.Add(time.Duration(s.cfg.Session.AbsoluteTimeoutHours) * time.Hour)

	_, err = s.db.Exec(ctx, `
		INSERT INTO auth_sessions (token_hash, actor_type, actor_id, created_at, last_activity_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, tokenHash, string(ActorTypeUser), email, now, now, absoluteExpiry)
	if err != nil {
		return Session{}, fmt.Errorf("insert session: %w", err)
	}

	return Session{
		Token: token,
		Actor: Actor{Type: ActorTypeUser, ID: email, Email: email},
	}, nil
}

func (s *SessionStore) ValidateToken(ctx context.Context, token string) (Actor, error) {
	tokenHash := hashToken(token)

	var actorType, actorID string
	var lastActivity, expiresAt time.Time
	err := s.db.QueryRow(ctx, `
		SELECT actor_type, actor_id, last_activity_at, expires_at
		FROM auth_sessions
		WHERE token_hash = $1
	`, tokenHash).Scan(&actorType, &actorID, &lastActivity, &expiresAt)
	if err != nil {
		return Actor{}, fmt.Errorf("session not found")
	}

	now := time.Now().UTC()
	if now.After(expiresAt) {
		_, _ = s.db.Exec(ctx, `DELETE FROM auth_sessions WHERE token_hash = $1`, tokenHash)
		return Actor{}, fmt.Errorf("session expired")
	}

	idleDeadline := lastActivity.Add(time.Duration(s.cfg.Session.IdleTimeoutMinutes) * time.Minute)
	if now.After(idleDeadline) {
		_, _ = s.db.Exec(ctx, `DELETE FROM auth_sessions WHERE token_hash = $1`, tokenHash)
		return Actor{}, fmt.Errorf("session idle timeout")
	}

	_, err = s.db.Exec(ctx, `
		UPDATE auth_sessions
		SET last_activity_at = $2
		WHERE token_hash = $1
	`, tokenHash, now)
	if err != nil {
		return Actor{}, fmt.Errorf("update session activity: %w", err)
	}

	actor := Actor{Type: ActorType(actorType), ID: actorID}
	if actor.Type == ActorTypeUser {
		actor.Email = actorID
	}
	return actor, nil
}

func (s *SessionStore) RevokeToken(ctx context.Context, token string) error {
	tokenHash := hashToken(token)
	_, err := s.db.Exec(ctx, `DELETE FROM auth_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (s *SessionStore) CleanupExpired(ctx context.Context) error {
	now := time.Now().UTC()
	_, err := s.db.Exec(ctx, `
		DELETE FROM auth_sessions
		WHERE expires_at < $1
		   OR last_activity_at < $2
	`, now, now.Add(-time.Duration(s.cfg.Session.IdleTimeoutMinutes)*time.Minute))
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `DELETE FROM auth_oauth_states WHERE expires_at < $1`, now)
	return err
}

func newSessionToken() (token string, tokenHash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(buf)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
