package auth

import (
	"fmt"
	"os"
	"strings"

	"config-manager/internal/config"
)

const (
	apiKeyPrefix     = "cm_live_"
	defaultCookie    = "cm_session"
	oauthStateTTLMin = 10
)

// Config holds authentication settings loaded from application.yaml and env.
type Config struct {
	Enabled             bool
	AllowedEmailDomains []string
	Google              GoogleConfig
	Session             SessionConfig
	APIKeys             APIKeysConfig
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	HomeDomain   string
}

type SessionConfig struct {
	CookieName           string
	CookieSecure         bool
	IdleTimeoutMinutes   int
	AbsoluteTimeoutHours int
}

type APIKeysConfig struct {
	BootstrapFromEnv bool
}

// LoadConfig reads auth settings from config file and environment.
func LoadConfig() (Config, error) {
	cfg := Config{
		Enabled:             config.Bool("auth.enabled", false),
		AllowedEmailDomains: normalizeDomains(config.StringSlice("auth.allowedEmailDomains")),
		Google: GoogleConfig{
			ClientID:     firstNonEmpty(os.Getenv("GOOGLE_CLIENT_ID"), config.String("auth.google.clientId", "")),
			ClientSecret: firstNonEmpty(os.Getenv("GOOGLE_CLIENT_SECRET"), config.String("auth.google.clientSecret", "")),
			RedirectURL:  firstNonEmpty(os.Getenv("AUTH_REDIRECT_URL"), config.String("auth.google.redirectUrl", "")),
			HomeDomain:   firstNonEmpty(os.Getenv("AUTH_GOOGLE_HOME_DOMAIN"), config.String("auth.google.homeDomain", "")),
		},
		Session: SessionConfig{
			CookieName:           firstNonEmpty(config.String("auth.session.cookieName", ""), defaultCookie),
			CookieSecure:         config.Bool("auth.session.cookieSecure", true),
			IdleTimeoutMinutes:   config.Int("auth.session.idleTimeoutMinutes", 480),
			AbsoluteTimeoutHours: config.Int("auth.session.absoluteTimeoutHours", 24),
		},
		APIKeys: APIKeysConfig{
			BootstrapFromEnv: config.Bool("auth.apiKeys.bootstrapFromEnv", true),
		},
	}

	if !cfg.Enabled {
		return cfg, nil
	}

	if cfg.Google.ClientID == "" || cfg.Google.ClientSecret == "" {
		return Config{}, fmt.Errorf("auth enabled but GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are required")
	}
	if cfg.Google.RedirectURL == "" {
		return Config{}, fmt.Errorf("auth enabled but auth.google.redirectUrl (or AUTH_REDIRECT_URL) is required")
	}
	if len(cfg.AllowedEmailDomains) == 0 {
		return Config{}, fmt.Errorf("auth enabled but auth.allowedEmailDomains must contain at least one domain")
	}
	if cfg.Session.IdleTimeoutMinutes <= 0 {
		return Config{}, fmt.Errorf("auth.session.idleTimeoutMinutes must be > 0")
	}
	if cfg.Session.AbsoluteTimeoutHours <= 0 {
		return Config{}, fmt.Errorf("auth.session.absoluteTimeoutHours must be > 0")
	}

	return cfg, nil
}

func normalizeDomains(domains []string) []string {
	seen := make(map[string]struct{}, len(domains))
	out := make([]string, 0, len(domains))
	for _, d := range domains {
		d = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(d, "@")))
		if d == "" {
			continue
		}
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		out = append(out, d)
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
