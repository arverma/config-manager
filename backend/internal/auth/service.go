package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service coordinates authentication for the API.
type Service struct {
	cfg       Config
	db        *pgxpool.Pool
	sessions  *SessionStore
	apiKeys   *APIKeyStore
	google    *GoogleAuth
}

func NewService(db *pgxpool.Pool) (*Service, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	svc := &Service{
		cfg:      cfg,
		db:       db,
		sessions: NewSessionStore(db, cfg),
		apiKeys:  NewAPIKeyStore(db),
	}

	if cfg.Enabled {
		svc.google = NewGoogleAuth(cfg, db, svc.sessions)
		if cfg.APIKeys.BootstrapFromEnv {
			if err := svc.apiKeys.BootstrapFromEnv(context.Background()); err != nil {
				return nil, err
			}
		}
	}

	return svc, nil
}

func (s *Service) Enabled() bool {
	return s.cfg.Enabled
}

func (s *Service) Config() Config {
	return s.cfg
}

func (s *Service) CreateAPIKey(ctx context.Context, name string) (string, error) {
	return s.apiKeys.CreateKey(ctx, name)
}

func (s *Service) ResolveRequest(ctx context.Context, r *http.Request) (Actor, error) {
	if authHeader := strings.TrimSpace(r.Header.Get("Authorization")); authHeader != "" {
		const prefix = "Bearer "
		if strings.HasPrefix(authHeader, prefix) {
			token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
			if token != "" {
				return s.apiKeys.ValidateKey(ctx, token)
			}
		}
	}

	cookie, err := r.Cookie(s.cfg.Session.CookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return Actor{}, fmt.Errorf("unauthenticated")
	}
	return s.sessions.ValidateToken(ctx, cookie.Value)
}

func (s *Service) SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.Session.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.Session.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   s.cfg.Session.AbsoluteTimeoutHours * 3600,
	})
}

func (s *Service) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.Session.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.Session.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	})
}

func (s *Service) SessionCookieName() string {
	return s.cfg.Session.CookieName
}

func (s *Service) GoogleLoginURL(ctx context.Context, returnTo string) (string, error) {
	return s.google.LoginURL(ctx, returnTo)
}

func (s *Service) GoogleCallback(ctx context.Context, code, state string) (Session, string, error) {
	return s.google.HandleCallback(ctx, code, state)
}

func (s *Service) RevokeSession(ctx context.Context, r *http.Request) error {
	cookie, err := r.Cookie(s.cfg.Session.CookieName)
	if err != nil {
		return nil
	}
	return s.sessions.RevokeToken(ctx, cookie.Value)
}

func (s *Service) CleanupExpired(ctx context.Context) {
	_ = s.sessions.CleanupExpired(ctx)
}

func IsPublicPath(path string) bool {
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}
	switch path {
	case "/healthz", "/readyz", "/auth/login/google", "/auth/callback/google", "/auth/session", "/auth/logout":
		return true
	default:
		return false
	}
}

func NormalizeRoutePath(r *http.Request, basePath string) string {
	path := r.URL.Path
	basePath = strings.TrimSuffix(strings.TrimSpace(basePath), "/")
	if basePath != "" && strings.HasPrefix(path, basePath) {
		path = strings.TrimPrefix(path, basePath)
		if path == "" {
			path = "/"
		}
	}
	return path
}
