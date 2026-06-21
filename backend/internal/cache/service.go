package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Service wraps an optional Redis client for config caching and auth backends.
type Service struct {
	cfg Config
	rdb *redis.Client
}

// New creates a cache service. When cache.enabled=false, Redis is not connected.
func New() (*Service, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return &Service{cfg: cfg}, nil
	}

	opts, err := redisOptions(cfg)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Service{cfg: cfg, rdb: rdb}, nil
}

func redisOptions(cfg Config) (*redis.Options, error) {
	if cfg.RedisURL != "" {
		opts, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("parse REDIS_URL: %w", err)
		}
		return opts, nil
	}

	return &redis.Options{
		Addr:     redisAddr(cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPass,
	}, nil
}

func (s *Service) Enabled() bool {
	return s.cfg.Enabled
}

func (s *Service) Config() Config {
	return s.cfg
}

func (s *Service) Redis() *redis.Client {
	return s.rdb
}

func (s *Service) KeyPrefix() string {
	return s.cfg.KeyPrefix
}

func (s *Service) Ping(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}
	return s.rdb.Ping(ctx).Err()
}

func (s *Service) Close() error {
	if s.rdb == nil {
		return nil
	}
	return s.rdb.Close()
}
