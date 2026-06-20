package auth

import (
	"net/http/httptest"
	"testing"
)

func TestIsPublicPath(t *testing.T) {
	public := []string{
		"/healthz",
		"/readyz",
		"/auth/login/google",
		"/auth/callback/google",
		"/auth/session",
		"/auth/logout",
	}
	for _, path := range public {
		if !IsPublicPath(path) {
			t.Fatalf("expected public path %q", path)
		}
	}
	if IsPublicPath("/namespaces") {
		t.Fatal("expected /namespaces to require auth when enabled")
	}
}

func TestNormalizeRoutePath(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/api/namespaces", nil)
	if got := NormalizeRoutePath(req, "/api"); got != "/namespaces" {
		t.Fatalf("got %q", got)
	}

	req = httptest.NewRequest("GET", "http://example.com/namespaces", nil)
	if got := NormalizeRoutePath(req, ""); got != "/namespaces" {
		t.Fatalf("got %q", got)
	}
}
