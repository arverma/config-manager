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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleAuth struct {
	cfg         Config
	oauth       *oauth2.Config
	sessions    SessionStore
	oauthStates OAuthStateStore
}

func NewGoogleAuth(cfg Config, sessions SessionStore, oauthStates OAuthStateStore) *GoogleAuth {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &GoogleAuth{cfg: cfg, oauth: oauthCfg, sessions: sessions, oauthStates: oauthStates}
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
	returnTo, err := g.oauthStates.ConsumeOAuthState(ctx, state)
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

	session, err := g.sessions.CreateUserSession(ctx, email)
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

	if err := g.oauthStates.CreateOAuthState(ctx, state, returnTo, expiresAt); err != nil {
		return "", err
	}
	return state, nil
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
