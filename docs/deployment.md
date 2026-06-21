# Deployment

## Health endpoints

- `GET /healthz`: liveness probe (when `HTTP_BASE_PATH` is empty)
- `GET /readyz`: readiness probe (when `HTTP_BASE_PATH` is empty)

When `HTTP_BASE_PATH=/api` (recommended for single-domain, path-based ingress):

- `GET /api/healthz`
- `GET /api/readyz` (checks Postgres reachability; checks Redis when `cache.enabled=true`)

## Path-based routing (recommended)

When deploying behind a single domain, route:

- `/` → UI
- `/api/*` → API

In this mode, the UI calls the API using same-origin paths (e.g. `fetch("/api/namespaces")`).

## Production checklist

- Database
  - Backups configured
  - Use TLS/SSL in production (`DATABASE_URL` with `sslmode=require` or set `DB_SSLMODE=require`)
- API
  - `CORS_ALLOWED_ORIGINS` restricted to your UI domains
  - Health checks wired to your platform (k8s, ECS, etc.)
  - Logs collected centrally
- UI
  - Prefer path-based routing so the browser can call `/api/*` (no public env needed).
- Authentication (production)
  - Enable `auth.enabled` in Helm; store Google OAuth credentials and API keys in Secrets.
  - Register OAuth redirect URI: `https://<host>/api/auth/callback/google`
  - Restrict access with `auth.allowedEmailDomains`
  - See [auth.md](auth.md)
- Caching (production, optional)
  - Enable `cache.enabled` for multi-replica API deployments with external Redis.
  - Store `REDIS_PASSWORD` (or `REDIS_URL`) in a Secret; set `cache.redis.host` / `port` in values (no in-chart Redis).
  - See [caching.md](caching.md)

