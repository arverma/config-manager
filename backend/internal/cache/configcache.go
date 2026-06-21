package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *Service) latestConfigKey(namespace, path string) string {
	return fmt.Sprintf("%scfg:%s:%s:latest", s.cfg.KeyPrefix, namespace, path)
}

// GetLatestConfig returns cached JSON for GET latest config, or ok=false on miss.
func (s *Service) GetLatestConfig(ctx context.Context, namespace, path string) ([]byte, bool, error) {
	if !s.Enabled() {
		return nil, false, nil
	}
	val, err := s.rdb.Get(ctx, s.latestConfigKey(namespace, path)).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}

// SetLatestConfig stores serialized GET latest config response JSON.
func (s *Service) SetLatestConfig(ctx context.Context, namespace, path string, payload []byte) error {
	if !s.Enabled() {
		return nil
	}
	ttl := time.Duration(s.cfg.ConfigTTL) * time.Second
	return s.rdb.Set(ctx, s.latestConfigKey(namespace, path), payload, ttl).Err()
}

// InvalidateLatestConfig removes cached latest config for the given identity.
func (s *Service) InvalidateLatestConfig(ctx context.Context, namespace, path string) error {
	if !s.Enabled() {
		return nil
	}
	namespace = strings.TrimSpace(namespace)
	path = strings.TrimSpace(path)
	if namespace == "" || path == "" {
		return nil
	}
	return s.rdb.Del(ctx, s.latestConfigKey(namespace, path)).Err()
}
