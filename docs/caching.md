# Redis caching (optional)

Redis-backed caching is **disabled by default**, similar to authentication. Local development uses Postgres only (`make db-up` / `docker-compose.yaml` has no Redis).

When `cache.enabled=true`, the API requires a reachable **external Redis** instance (no Bitnami subchart). Startup and `/readyz` fail if Redis is misconfigured or unreachable.

## What Redis is used for

| Feature | When cache enabled |
|---------|-------------------|
| `GET /configs/{namespace}/{path}` (latest) | Cache-aside; invalidated on create/update/delete |
| Browser sessions (`auth.enabled=true`) | Stored in Redis (not Postgres `auth_sessions`) |
| OAuth CSRF state | Stored in Redis (not Postgres `auth_oauth_states`) |
| API keys | Always Postgres (no Redis cache in v1) |

Enabling cache in production **logs out existing browser sessions** once (Postgres sessions are not migrated).

## Configuration

### application.yaml

```yaml
cache:
  enabled: false
  redis:
    host: ""
    port: 6379
    keyPrefix: "cm:"
  config:
    ttlSeconds: 3600
```

### Environment variables

When cache is enabled, provide either:

- `REDIS_URL` (full connection string, takes precedence), or
- `REDIS_HOST`, `REDIS_PORT` (default `6379`), and `REDIS_PASSWORD`

Override config file keys with `CONFIG_MANAGER_CACHE_*` (e.g. `CONFIG_MANAGER_CACHE_ENABLED=true`).

### Helm values

```yaml
cache:
  enabled: true
  redis:
    host: "10.0.0.5"              # external Redis IP/hostname
    port: 6379
    existingSecretName: config-manager-redis   # REDIS_PASSWORD
    existingSecretKey: "REDIS_PASSWORD"
    keyPrefix: "cm:"
  config:
    ttlSeconds: 3600
```

Password must **not** appear in plain Helm values. Options:

1. **Dedicated secret** — `kubectl create secret generic config-manager-redis --from-literal=REDIS_PASSWORD='...'`
2. **ESO / Vault** — add `REDIS_PASSWORD` to the same Vault path synced by `api.externalSecret` (alongside `DB_PASSWORD`)

Pattern mirrors Superset-style external Redis: set `cache.redis.host` / `port` in values; credentials from a Secret.

## Redis key layout

| Key | Purpose | TTL |
|-----|---------|-----|
| `{prefix}cfg:{namespace}:{path}:latest` | Serialized GET latest response | `cache.config.ttlSeconds` (invalidated on writes) |
| `{prefix}sess:{token_hash}` | Session record JSON | Absolute session expiry |
| `{prefix}oauth:{state}` | OAuth return path | 10 minutes |

Default prefix: `cm:`.

## Multi-replica deployments

Shared Redis gives **consistent cache and sessions** across API pods. You can scale API replicas horizontally when cache is enabled.

## Failure behavior

When `cache.enabled=true`, Redis is **hard-required**:

- Process exits at startup if Redis cannot be pinged
- `/readyz` returns `503` if Redis becomes unreachable

There is no fallback to Postgres for sessions or config cache when cache mode is on.

## Local development

Leave `cache.enabled=false`. No Redis container is required.

To test Redis locally, run Redis yourself and set `CONFIG_MANAGER_CACHE_ENABLED=true` plus `REDIS_HOST` / `REDIS_PASSWORD`.

## v2 (not yet implemented)

- ETag / `If-None-Match` using `content_sha256`
- Cache GET specific version, list, and browse endpoints
- API key validation cache in Redis

See [roadmap.md](roadmap.md).
