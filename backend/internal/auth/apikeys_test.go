package auth

import "testing"

func TestHashAPIKeyStable(t *testing.T) {
	raw := "cm_live_abc123"
	a := hashAPIKey(raw)
	b := hashAPIKey(raw)
	if a != b {
		t.Fatalf("hash not stable")
	}
	if a == raw {
		t.Fatal("hash should not equal raw key")
	}
}
