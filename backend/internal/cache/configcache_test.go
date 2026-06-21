package cache

import (
	"context"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestConfigCacheHitMissInvalidate(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{
		cfg: Config{
			Enabled:   true,
			KeyPrefix: "cm:",
			ConfigTTL: 3600,
		},
		rdb: rdb,
	}

	ctx := context.Background()
	payload := []byte(`{"config":{"namespace":"ns","path":"app.yaml"},"latest":{"version":1}}`)

	got, ok, err := svc.GetLatestConfig(ctx, "ns", "app.yaml")
	if err != nil {
		t.Fatalf("get miss: %v", err)
	}
	if ok || got != nil {
		t.Fatalf("expected miss, got ok=%v data=%q", ok, got)
	}

	if err := svc.SetLatestConfig(ctx, "ns", "app.yaml", payload); err != nil {
		t.Fatalf("set: %v", err)
	}

	got, ok, err = svc.GetLatestConfig(ctx, "ns", "app.yaml")
	if err != nil {
		t.Fatalf("get hit: %v", err)
	}
	if !ok || string(got) != string(payload) {
		t.Fatalf("got ok=%v data=%q", ok, got)
	}

	if err := svc.InvalidateLatestConfig(ctx, "ns", "app.yaml"); err != nil {
		t.Fatalf("invalidate: %v", err)
	}

	got, ok, err = svc.GetLatestConfig(ctx, "ns", "app.yaml")
	if err != nil {
		t.Fatalf("get after invalidate: %v", err)
	}
	if ok || got != nil {
		t.Fatalf("expected miss after invalidate, got ok=%v", ok)
	}
}

func TestDisabledCacheNoOps(t *testing.T) {
	svc := &Service{cfg: Config{Enabled: false}}
	ctx := context.Background()

	if _, ok, err := svc.GetLatestConfig(ctx, "ns", "p"); err != nil || ok {
		t.Fatalf("get disabled: ok=%v err=%v", ok, err)
	}
	if err := svc.SetLatestConfig(ctx, "ns", "p", []byte("{}")); err != nil {
		t.Fatalf("set disabled: %v", err)
	}
	if err := svc.InvalidateLatestConfig(ctx, "ns", "p"); err != nil {
		t.Fatalf("invalidate disabled: %v", err)
	}
	if err := svc.Ping(ctx); err != nil {
		t.Fatalf("ping disabled: %v", err)
	}
}
