package httpapi

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type ConfigFormat string

const (
	FormatJSON       ConfigFormat = "json"
	FormatYAML       ConfigFormat = "yaml"
	maxCursorOffset               = 100_000
)

func parseOptionalBool(req *http.Request, key string) (bool, bool, error) {
	raw := strings.TrimSpace(req.URL.Query().Get(key))
	if raw == "" {
		return false, false, nil
	}
	switch strings.ToLower(raw) {
	case "true", "1", "yes", "y":
		return true, true, nil
	case "false", "0", "no", "n":
		return false, true, nil
	default:
		return false, true, fmt.Errorf("%s must be a boolean", key)
	}
}

func parseLimit(req *http.Request, def int) (int, error) {
	raw := strings.TrimSpace(req.URL.Query().Get("limit"))
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > 500 {
		return 0, fmt.Errorf("limit must be an integer between 1 and 500")
	}
	return n, nil
}

// Cursor is an opaque string. We implement it as base64("o:<offset>").
func parseCursorOffset(req *http.Request) (int, error) {
	raw := strings.TrimSpace(req.URL.Query().Get("cursor"))
	if raw == "" {
		return 0, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		n, err2 := strconv.Atoi(raw)
		if err2 != nil || n < 0 {
			return 0, fmt.Errorf("invalid cursor")
		}
		if n > maxCursorOffset {
			return 0, fmt.Errorf("cursor offset exceeds maximum")
		}
		return n, nil
	}
	s := string(decoded)
	if !strings.HasPrefix(s, "o:") {
		return 0, fmt.Errorf("invalid cursor")
	}
	n, err := strconv.Atoi(strings.TrimPrefix(s, "o:"))
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid cursor")
	}
	if n > maxCursorOffset {
		return 0, fmt.Errorf("cursor offset exceeds maximum")
	}
	return n, nil
}

func encodeCursorOffset(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("o:%d", offset)))
}

func parsePrefix(req *http.Request) (string, error) {
	prefix := strings.TrimSpace(req.URL.Query().Get("prefix"))
	if prefix == "" {
		return "", nil
	}
	prefix = strings.TrimPrefix(prefix, "/")
	if strings.Contains(prefix, "..") {
		return "", fmt.Errorf("prefix must not contain '..'")
	}
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return prefix, nil
}
