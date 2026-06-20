package auth

import "testing"

func TestEmailAllowed(t *testing.T) {
	allowed := []string{"example.com", "partner.org"}

	cases := []struct {
		email string
		want  bool
	}{
		{"user@example.com", true},
		{"USER@EXAMPLE.COM", true},
		{"bot@partner.org", true},
		{"user@gmail.com", false},
		{"not-an-email", false},
		{"@example.com", false},
		{"user@", false},
	}

	for _, tc := range cases {
		if got := EmailAllowed(tc.email, allowed); got != tc.want {
			t.Fatalf("EmailAllowed(%q) = %v, want %v", tc.email, got, tc.want)
		}
	}
}

func TestValidateEmailAllowlist(t *testing.T) {
	err := ValidateEmailAllowlist("user@blocked.com", []string{"example.com"})
	if err == nil {
		t.Fatal("expected error for blocked domain")
	}
}
