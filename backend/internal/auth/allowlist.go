package auth

import (
	"fmt"
	"strings"
)

// EmailAllowed returns true when email belongs to one of the configured domains.
func EmailAllowed(email string, allowedDomains []string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return false
	}
	domain := email[at+1:]
	for _, allowed := range allowedDomains {
		if domain == strings.ToLower(strings.TrimSpace(allowed)) {
			return true
		}
	}
	return false
}

// ValidateEmailAllowlist returns an error when email is not allowed.
func ValidateEmailAllowlist(email string, allowedDomains []string) error {
	if EmailAllowed(email, allowedDomains) {
		return nil
	}
	return fmt.Errorf("email domain is not allowed")
}
