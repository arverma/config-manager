package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/url"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"config-manager/internal/httpapi"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	port := getenvDefault("PORT", "8080")
	databaseURL, err := databaseURLFromEnv()
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewRouter(pool),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MiB
	}

	go func() {
		log.Printf("config-manager listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)
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
