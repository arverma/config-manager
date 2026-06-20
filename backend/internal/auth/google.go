package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleAuth struct {
	cfg    Config
	db     *pgxpool.Pool
	oauth  *oauth2.Config
	store  *SessionStore
}

func NewGoogleAuth(cfg Config, db *pgxpool.Pool, sessions *SessionStore) *GoogleAuth {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &GoogleAuth{cfg: cfg, db: db, oauth: oauthCfg, store: sessions}
}

func (g *GoogleAuth) LoginURL(ctx context.Context, returnTo string) (string, error) {
	state, err := g.createOAuthState(ctx, returnTo)
	if err != nil {
		return "", err
	}

	opts := []oauth2.AuthCodeOption{}
	if hd := strings.TrimSpace(g.cfg.Google.HomeDomain); hd != "" {
		opts = append(opts, oauth2.SetAuthURLParam("hd", hd))
	}

	return g.oauth.AuthCodeURL(state, opts...), nil
}

func (g *GoogleAuth) HandleCallback(ctx context.Context, code, state string) (Session, string, error) {
	returnTo, err := g.consumeOAuthState(ctx, state)
	if err != nil {
		return Session{}, "", err
	}

	token, err := g.oauth.Exchange(ctx, code)
	if err != nil {
		return Session{}, "", fmt.Errorf("oauth exchange: %w", err)
	}

	email, err := g.fetchGoogleEmail(ctx, token.AccessToken)
	if err != nil {
		return Session{}, "", err
	}

	if err := ValidateEmailAllowlist(email, g.cfg.AllowedEmailDomains); err != nil {
		return Session{}, "", err
	}

	session, err := g.store.CreateUserSession(ctx, email)
	if err != nil {
		return Session{}, "", err
	}

	return session, returnTo, nil
}

func (g *GoogleAuth) createOAuthState(ctx context.Context, returnTo string) (string, error) {
	if returnTo == "" {
		returnTo = "/"
	}
	if !strings.HasPrefix(returnTo, "/") || strings.HasPrefix(returnTo, "//") {
		returnTo = "/"
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	state := base64.RawURLEncoding.EncodeToString(buf)
	expiresAt := time.Now().UTC().Add(oauthStateTTLMin * time.Minute)

	_, err := g.db.Exec(ctx, `
		INSERT INTO auth_oauth_states (state, return_to, expires_at)
		VALUES ($1, $2, $3)
	`, state, returnTo, expiresAt)
	if err != nil {
		return "", fmt.Errorf("store oauth state: %w", err)
	}
	return state, nil
}

func (g *GoogleAuth) consumeOAuthState(ctx context.Context, state string) (string, error) {
	state = strings.TrimSpace(state)
	if state == "" {
		return "", fmt.Errorf("missing oauth state")
	}

	var returnTo string
	var expiresAt time.Time
	err := g.db.QueryRow(ctx, `
		SELECT return_to, expires_at
		FROM auth_oauth_states
		WHERE state = $1
	`, state).Scan(&returnTo, &expiresAt)
	if err != nil {
		return "", fmt.Errorf("invalid oauth state")
	}

	_, _ = g.db.Exec(ctx, `DELETE FROM auth_oauth_states WHERE state = $1`, state)

	if time.Now().UTC().After(expiresAt) {
		return "", fmt.Errorf("oauth state expired")
	}
	if returnTo == "" {
		returnTo = "/"
	}
	return returnTo, nil
}

func (g *GoogleAuth) fetchGoogleEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("google userinfo: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google userinfo status %d", res.StatusCode)
	}

	var payload struct {
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("parse google userinfo: %w", err)
	}
	email := strings.TrimSpace(strings.ToLower(payload.Email))
	if email == "" {
		return "", fmt.Errorf("google account has no email")
	}
	if !payload.VerifiedEmail {
		return "", fmt.Errorf("google email is not verified")
	}
	return email, nil
}
