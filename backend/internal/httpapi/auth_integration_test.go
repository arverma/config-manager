package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"config-manager/internal/auth"
	"config-manager/internal/config"
	"config-manager/migrations"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRouterAuthDisabledNamespaces(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "application.yaml")
	cfg := []byte(`
auth:
  enabled: false
api:
  request:
    requestTimeoutSeconds: 30
`)
	if err := os.WriteFile(cfgPath, cfg, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := config.Load(cfgPath); err != nil {
		t.Fatalf("config load: %v", err)
	}

	pool := testPoolOrSkip(t)
	defer pool.Close()
	authSvc, err := auth.NewService(pool)
	if err != nil {
		t.Fatalf("auth service: %v", err)
	}
	if authSvc.Enabled() {
		t.Skip("test expects auth.enabled=false")
	}

	handler := NewRouter(pool, authSvc)
	req := httptest.NewRequest(http.MethodGet, "/namespaces", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRouterAuthEnabledRequiresCredentials(t *testing.T) {
	pool := testPoolOrSkip(t)
	defer pool.Close()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "application.yaml")
	cfg := []byte(`
auth:
  enabled: true
  allowedEmailDomains:
    - example.com
  google:
    clientId: test-client
    clientSecret: test-secret
    redirectUrl: http://localhost:8080/auth/callback/google
  session:
    cookieSecure: false
api:
  request:
    requestTimeoutSeconds: 30
`)
	if err := os.WriteFile(cfgPath, cfg, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := config.Load(cfgPath); err != nil {
		t.Fatalf("config load: %v", err)
	}

	authSvc, err := auth.NewService(pool)
	if err != nil {
		t.Fatalf("auth service: %v", err)
	}
	if !authSvc.Enabled() {
		t.Fatal("expected auth enabled")
	}

	handler := NewRouter(pool, authSvc)
	req := httptest.NewRequest(http.MethodGet, "/namespaces", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRouterAuthEnabledAPIKey(t *testing.T) {
	pool := testPoolOrSkip(t)
	defer pool.Close()
	ctx := context.Background()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "application.yaml")
	cfg := []byte(`
auth:
  enabled: true
  allowedEmailDomains:
    - example.com
  google:
    clientId: test-client
    clientSecret: test-secret
    redirectUrl: http://localhost:8080/auth/callback/google
  session:
    cookieSecure: false
api:
  request:
    requestTimeoutSeconds: 30
`)
	if err := os.WriteFile(cfgPath, cfg, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := config.Load(cfgPath); err != nil {
		t.Fatalf("config load: %v", err)
	}

	authSvc, err := auth.NewService(pool)
	if err != nil {
		t.Fatalf("auth service: %v", err)
	}

	store := auth.NewAPIKeyStore(pool)
	rawKey, err := store.CreateKey(ctx, "test-key-"+t.Name())
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}

	handler := NewRouter(pool, authSvc)
	req := httptest.NewRequest(http.MethodGet, "/namespaces", nil)
	req.Header.Set("Authorization", "Bearer "+rawKey)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestHealthzPublicWhenAuthEnabled(t *testing.T) {
	pool := testPoolOrSkip(t)
	defer pool.Close()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "application.yaml")
	cfg := []byte(`
auth:
  enabled: true
  allowedEmailDomains:
    - example.com
  google:
    clientId: test-client
    clientSecret: test-secret
    redirectUrl: http://localhost:8080/auth/callback/google
api:
  request:
    requestTimeoutSeconds: 30
`)
	if err := os.WriteFile(cfgPath, cfg, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := config.Load(cfgPath); err != nil {
		t.Fatalf("config load: %v", err)
	}

	authSvc, err := auth.NewService(pool)
	if err != nil {
		t.Fatalf("auth service: %v", err)
	}

	handler := NewRouter(pool, authSvc)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func testPoolOrSkip(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@127.0.0.1:5432/config_manager?sslmode=disable"
	}

	if err := runTestMigrations(databaseURL); err != nil {
		t.Skipf("migrate failed: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("postgres ping failed: %v", err)
	}
	return pool
}

func runTestMigrations(databaseURL string) error {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	migrateDBURL := strings.Replace(databaseURL, "postgres://", "pgx5://", 1)
	if migrateDBURL == databaseURL {
		migrateDBURL = "pgx5://" + strings.TrimPrefix(databaseURL, "postgres:")
	}
	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, migrateDBURL)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
