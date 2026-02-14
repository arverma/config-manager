package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"config-manager/internal/commons"
	"config-manager/internal/config"
	"config-manager/internal/httpapi"
	"config-manager/migrations"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	configPath := flag.String("config", "", "path to application.yaml (server, DB retry, timeouts); overridden by CONFIG_MANAGER_* env vars")
	flag.Parse()

	if err := config.Load(*configPath); err != nil {
		log.Fatalf("config: %v", err)
	}

	port := getenvDefault("PORT", "8080")
	databaseURL, err := databaseURLFromEnv()
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	maxAttempts := config.Int("api.databaseRetry.maxAttempts", 5)
	initialBackoff := time.Duration(config.Int("api.databaseRetry.retryBackoffSeconds", 2)) * time.Second

	var pool *pgxpool.Pool
	if err := commons.RetryWithBackoff(maxAttempts, initialBackoff, func() error {
		var connectErr error
		pool, connectErr = pgxpool.New(ctx, databaseURL)
		return connectErr
	}); err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	if err := commons.RetryWithBackoff(maxAttempts, initialBackoff, func() error {
		return runMigrations(databaseURL)
	}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	readHeaderTimeout := time.Duration(config.Int("api.server.readHeaderTimeoutSeconds", 5)) * time.Second
	readTimeout := time.Duration(config.Int("api.server.readTimeoutSeconds", 30)) * time.Second
	writeTimeout := time.Duration(config.Int("api.server.writeTimeoutSeconds", 30)) * time.Second
	idleTimeout := time.Duration(config.Int("api.server.idleTimeoutSeconds", 60)) * time.Second
	shutdownTimeout := time.Duration(config.Int("api.server.shutdownTimeoutSeconds", 10)) * time.Second

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewRouter(pool),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    1 << 20, // 1 MiB
	}

	go func() {
		log.Printf("config-manager listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)
}

func runMigrations(databaseURL string) error {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	// golang-migrate pgx v5 driver expects pgx5:// scheme
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

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func databaseURLFromEnv() (string, error) {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v, nil
	}

	host := strings.TrimSpace(os.Getenv("DB_HOST"))
	dbName := strings.TrimSpace(os.Getenv("DB_NAME"))
	user := strings.TrimSpace(os.Getenv("DB_USER"))
	password := os.Getenv("DB_PASSWORD")
	port := strings.TrimSpace(getenvDefault("DB_PORT", "5432"))

	if host == "" || dbName == "" || user == "" || password == "" {
		return "", errors.New("DATABASE_URL is required, or set DB_HOST, DB_PORT (optional), DB_NAME, DB_USER, DB_PASSWORD (and optionally DB_SSLMODE)")
	}

	sslmode := strings.TrimSpace(os.Getenv("DB_SSLMODE"))
	if sslmode == "" {
		// Safe-by-default for production; local dev can override with DB_SSLMODE=disable.
		sslmode = "require"
		if host == "localhost" || host == "127.0.0.1" {
			sslmode = "disable"
		}
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + dbName,
	}
	q := url.Values{}
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
