package auth

import (
	"context"
	"time"
)

// SessionStore persists browser session tokens.
type SessionStore interface {
	CreateUserSession(ctx context.Context, email string) (Session, error)
	ValidateToken(ctx context.Context, token string) (Actor, error)
	RevokeToken(ctx context.Context, token string) error
	CleanupExpired(ctx context.Context) error
	UsesRedis() bool
}

// OAuthStateStore persists short-lived OAuth CSRF state tokens.
type OAuthStateStore interface {
	CreateOAuthState(ctx context.Context, state, returnTo string, expiresAt time.Time) error
	ConsumeOAuthState(ctx context.Context, state string) (returnTo string, err error)
}
