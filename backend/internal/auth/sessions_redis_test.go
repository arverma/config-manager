package auth

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisSessionStoreLifecycle(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	cfg := Config{
		Session: SessionConfig{
			IdleTimeoutMinutes:   30,
			AbsoluteTimeoutHours: 24,
		},
	}
	store := NewRedisSessionStore(rdb, cfg, "cm:")
	ctx := context.Background()

	sess, err := store.CreateUserSession(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	actor, err := store.ValidateToken(ctx, sess.Token)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if actor.ID != "user@example.com" || actor.Email != "user@example.com" {
		t.Fatalf("actor = %+v", actor)
	}

	if err := store.RevokeToken(ctx, sess.Token); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := store.ValidateToken(ctx, sess.Token); err == nil {
		t.Fatal("expected session not found after revoke")
	}
}

func TestRedisSessionStoreIdleTimeout(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	cfg := Config{
		Session: SessionConfig{
			IdleTimeoutMinutes:   1,
			AbsoluteTimeoutHours: 24,
		},
	}
	store := NewRedisSessionStore(rdb, cfg, "cm:")
	ctx := context.Background()

	sess, err := store.CreateUserSession(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	tokenHash := hashToken(sess.Token)
	key := store.sessionKey(tokenHash)
	record := redisSessionRecord{
		ActorType:      string(ActorTypeUser),
		ActorID:        "user@example.com",
		LastActivityAt: time.Now().UTC().Add(-2 * time.Minute),
		ExpiresAt:      time.Now().UTC().Add(24 * time.Hour),
	}
	payload, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := rdb.Set(ctx, key, payload, time.Hour).Err(); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	if _, err := store.ValidateToken(ctx, sess.Token); err == nil {
		t.Fatal("expected idle timeout")
	}
}

func TestRedisOAuthStateConsumeOnce(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	store := NewRedisOAuthStateStore(rdb, "cm:")
	ctx := context.Background()
	expiresAt := time.Now().UTC().Add(10 * time.Minute)

	if err := store.CreateOAuthState(ctx, "state123", "/dashboard", expiresAt); err != nil {
		t.Fatalf("create: %v", err)
	}

	returnTo, err := store.ConsumeOAuthState(ctx, "state123")
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if returnTo != "/dashboard" {
		t.Fatalf("returnTo = %q", returnTo)
	}

	if _, err := store.ConsumeOAuthState(ctx, "state123"); err == nil {
		t.Fatal("expected invalid oauth state on second consume")
	}
}
