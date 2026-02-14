package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

func parseBody(format ConfigFormat, raw string) (any, []byte, error) {
	switch format {
	case FormatJSON:
		var anyVal any
		dec := json.NewDecoder(strings.NewReader(raw))
		dec.UseNumber()
		if err := dec.Decode(&anyVal); err != nil {
			return nil, nil, errors.New("invalid json")
		}
		if err := dec.Decode(&struct{}{}); err != io.EOF {
			return nil, nil, errors.New("invalid json")
		}
		j, _ := json.Marshal(anyVal)
		return anyVal, j, nil
	case FormatYAML:
		var anyVal any
		if err := yaml.Unmarshal([]byte(raw), &anyVal); err != nil {
			return nil, nil, errors.New("invalid yaml")
		}
		normalized := normalizeYAML(anyVal)
		j, _ := json.Marshal(normalized)
		return normalized, j, nil
	default:
		return nil, nil, errors.New("unknown format")
	}
}

func normalizeYAML(v any) any {
	switch t := v.(type) {
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, vv := range t {
			m[asString(k)] = normalizeYAML(vv)
		}
		return m
	case []any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = normalizeYAML(t[i])
		}
		return out
	default:
		return t
	}
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
