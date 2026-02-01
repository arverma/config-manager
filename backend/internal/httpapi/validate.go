package httpapi

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var namespaceRE = regexp.MustCompile(`^[a-z_]+$`)

func validateNamespace(namespace string) error {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return errors.New("namespace is required")
	}
	if !namespaceRE.MatchString(namespace) {
		return errors.New("namespace must match ^[a-z_]+$")
	}
	return nil
}

func validateNamespaceName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name is required")
	}
	if !namespaceRE.MatchString(name) {
		return errors.New("name must match ^[a-z_]+$")
	}
	return nil
}

func normalizeConfigPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("path is required")
	}

	// Keep existing behavior lenient for leading/trailing slashes.
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		return "", errors.New("path is required")
	}
	if strings.Contains(path, "//") {
		return "", errors.New("path must not contain empty segments")
	}
	// Keep it simple: reject any parent traversal marker.
	// (DB constraint is stricter; this catches the common cases early.)
	if strings.Contains(path, "..") {
		return "", errors.New("path must not contain '..'")
	}
	for _, r := range path {
		if unicode.IsSpace(r) {
			return "", errors.New("path must not contain whitespace")
		}
	}
	return path, nil
}
