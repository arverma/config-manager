package cache

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"config-manager/internal/config"
)

// Config holds Redis cache settings from application.yaml and environment.
type Config struct {
	Enabled     bool
	RedisHost   string
	RedisPort   int
	RedisURL    string
	RedisPass   string
	KeyPrefix   string
	ConfigTTL   int
}

// LoadConfig reads cache settings. When disabled, no Redis connection details are required.
func LoadConfig() (Config, error) {
	cfg := Config{
		Enabled:   config.Bool("cache.enabled", false),
		RedisHost: firstNonEmpty(os.Getenv("REDIS_HOST"), config.String("cache.redis.host", "")),
		RedisPort: config.Int("cache.redis.port", 6379),
		RedisURL:  strings.TrimSpace(os.Getenv("REDIS_URL")),
		RedisPass: strings.TrimSpace(os.Getenv("REDIS_PASSWORD")),
		KeyPrefix: firstNonEmpty(config.String("cache.redis.keyPrefix", ""), "cm:"),
		ConfigTTL: config.Int("cache.config.ttlSeconds", 3600),
	}

	if !cfg.Enabled {
		return cfg, nil
	}

	if cfg.RedisURL == "" && strings.TrimSpace(cfg.RedisHost) == "" {
		return Config{}, fmt.Errorf("cache enabled but REDIS_URL or cache.redis.host (REDIS_HOST) is required")
	}
	if cfg.ConfigTTL <= 0 {
		return Config{}, fmt.Errorf("cache.config.ttlSeconds must be > 0")
	}

	return cfg, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func redisAddr(host string, port int) string {
	return fmt.Sprintf("%s:%s", host, strconv.Itoa(port))
}
