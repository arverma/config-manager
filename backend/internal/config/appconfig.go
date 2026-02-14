package config

import (
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const envPrefix = "CONFIG_MANAGER_"
const defaultConfigPath = "confs/application.yaml"

// Load reads the YAML config file at path and stores it for Int() lookups.
// If path is empty, path defaults to "confs/application.yaml". After loading,
// CONFIG_MANAGER_* environment variables are applied as overrides (e.g. CONFIG_MANAGER_API_SERVER_READ_HEADER_TIMEOUT_SECONDS=5).
func Load(path string) error {
	if path == "" {
		path = defaultConfigPath
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}
	store = raw
	overrideWithEnv()
	return nil
}

var store map[string]any

// overrideWithEnv applies CONFIG_MANAGER_* env vars to the config store.
// CONFIG_MANAGER_API_SERVER_READ_HEADER_TIMEOUT_SECONDS=5 becomes api.server.readHeaderTimeoutSeconds = "5".
func overrideWithEnv() {
	if store == nil {
		return
	}
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, envPrefix) {
			continue
		}
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		envKey := parts[0][len(envPrefix):]
		envValue := parts[1]
		configKey := strings.ToLower(strings.ReplaceAll(envKey, "_", "."))
		setNested(store, configKey, envValue)
	}
}

// setNested sets a dot-separated key (e.g. "api.server.x") in the map, creating nested maps as needed.
func setNested(m map[string]any, key string, value string) {
	parts := strings.Split(key, ".")
	if len(parts) == 0 {
		return
	}
	cur := m
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, ok := cur[part].(map[string]any)
		if !ok || next == nil {
			next = make(map[string]any)
			cur[part] = next
		}
		cur = next
	}
	cur[parts[len(parts)-1]] = value
}

// Int returns the integer at the given dot-separated key (e.g. "api.server.readHeaderTimeoutSeconds").
// If the key is missing or not a number, returns defaultVal. String values are parsed as int.
func Int(key string, defaultVal int) int {
	if store == nil {
		return defaultVal
	}
	v, ok := getNested(store, key)
	if !ok {
		return defaultVal
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(n)); err == nil {
			return parsed
		}
		return defaultVal
	default:
		return defaultVal
	}
}

func getNested(m map[string]any, key string) (any, bool) {
	var current any = m
	start := 0
	for i := 0; i <= len(key); i++ {
		if i == len(key) || key[i] == '.' {
			if start == i {
				break
			}
			part := key[start:i]
			start = i + 1
			mm, ok := current.(map[string]any)
			if !ok {
				return nil, false
			}
			current, ok = mm[part]
			if !ok {
				return nil, false
			}
			if i == len(key) {
				return current, true
			}
		}
	}
	return nil, false
}
