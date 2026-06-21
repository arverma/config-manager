package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisOAuthStateStore struct {
	rdb    *redis.Client
	prefix string
}

func NewRedisOAuthStateStore(rdb *redis.Client, keyPrefix string) *RedisOAuthStateStore {
	return &RedisOAuthStateStore{rdb: rdb, prefix: keyPrefix}
}

func (s *RedisOAuthStateStore) oauthKey(state string) string {
	return fmt.Sprintf("%soauth:%s", s.prefix, state)
}

func (s *RedisOAuthStateStore) CreateOAuthState(ctx context.Context, state, returnTo string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = oauthStateTTLMin * time.Minute
	}
	return s.rdb.Set(ctx, s.oauthKey(state), returnTo, ttl).Err()
}

func (s *RedisOAuthStateStore) ConsumeOAuthState(ctx context.Context, state string) (string, error) {
	state = strings.TrimSpace(state)
	if state == "" {
		return "", fmt.Errorf("missing oauth state")
	}

	returnTo, err := s.rdb.GetDel(ctx, s.oauthKey(state)).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("invalid oauth state")
	}
	if err != nil {
		return "", fmt.Errorf("read oauth state: %w", err)
	}
	if returnTo == "" {
		returnTo = "/"
	}
	return returnTo, nil
}

type RedisSessionStore struct {
	rdb    *redis.Client
	cfg    Config
	prefix string
}

func NewRedisSessionStore(rdb *redis.Client, cfg Config, keyPrefix string) *RedisSessionStore {
	return &RedisSessionStore{rdb: rdb, cfg: cfg, prefix: keyPrefix}
}

func (s *RedisSessionStore) UsesRedis() bool { return true }

func (s *RedisSessionStore) sessionKey(tokenHash string) string {
	return fmt.Sprintf("%ssess:%s", s.prefix, tokenHash)
}

type redisSessionRecord struct {
	ActorType      string    `json:"actor_type"`
	ActorID        string    `json:"actor_id"`
	LastActivityAt time.Time `json:"last_activity_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

func (s *RedisSessionStore) CreateUserSession(ctx context.Context, email string) (Session, error) {
	token, tokenHash, err := newSessionToken()
	if err != nil {
		return Session{}, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(s.cfg.Session.AbsoluteTimeoutHours) * time.Hour)
	record := redisSessionRecord{
		ActorType:      string(ActorTypeUser),
		ActorID:        email,
		LastActivityAt: now,
		ExpiresAt:      expiresAt,
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return Session{}, err
	}

	ttl := time.Until(expiresAt)
	if err := s.rdb.Set(ctx, s.sessionKey(tokenHash), payload, ttl).Err(); err != nil {
		return Session{}, fmt.Errorf("store session: %w", err)
	}

	return Session{
		Token: token,
		Actor: Actor{Type: ActorTypeUser, ID: email, Email: email},
	}, nil
}

func (s *RedisSessionStore) ValidateToken(ctx context.Context, token string) (Actor, error) {
	tokenHash := hashToken(token)
	raw, err := s.rdb.Get(ctx, s.sessionKey(tokenHash)).Bytes()
	if err == redis.Nil {
		return Actor{}, fmt.Errorf("session not found")
	}
	if err != nil {
		return Actor{}, fmt.Errorf("read session: %w", err)
	}

	var record redisSessionRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		return Actor{}, fmt.Errorf("parse session: %w", err)
	}

	now := time.Now().UTC()
	if now.After(record.ExpiresAt) {
		_, _ = s.rdb.Del(ctx, s.sessionKey(tokenHash)).Result()
		return Actor{}, fmt.Errorf("session expired")
	}

	idleDeadline := record.LastActivityAt.Add(time.Duration(s.cfg.Session.IdleTimeoutMinutes) * time.Minute)
	if now.After(idleDeadline) {
		_, _ = s.rdb.Del(ctx, s.sessionKey(tokenHash)).Result()
		return Actor{}, fmt.Errorf("session idle timeout")
	}

	record.LastActivityAt = now
	payload, err := json.Marshal(record)
	if err != nil {
		return Actor{}, fmt.Errorf("marshal session: %w", err)
	}
	ttl := time.Until(record.ExpiresAt)
	if err := s.rdb.Set(ctx, s.sessionKey(tokenHash), payload, ttl).Err(); err != nil {
		return Actor{}, fmt.Errorf("update session activity: %w", err)
	}

	actor := Actor{Type: ActorType(record.ActorType), ID: record.ActorID}
	if actor.Type == ActorTypeUser {
		actor.Email = record.ActorID
	}
	return actor, nil
}

func (s *RedisSessionStore) RevokeToken(ctx context.Context, token string) error {
	tokenHash := hashToken(token)
	return s.rdb.Del(ctx, s.sessionKey(tokenHash)).Err()
}

func (s *RedisSessionStore) CleanupExpired(ctx context.Context) error {
	return nil
}
