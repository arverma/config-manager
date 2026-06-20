# Environment variables

## Backend (API)

Server timeouts, request timeout, readiness ping timeout, and DB retry settings are read from the application config file (e.g. `backend/confs/application.yaml` locally; in Helm the chart ships a default and the deployment repo can override per environment). They can be overridden at runtime by `CONFIG_MANAGER_*` env vars (e.g. `CONFIG_MANAGER_API_SERVER_READ_HEADER_TIMEOUT_SECONDS=5`).

### Required

- `DATABASE_URL`: Postgres connection string

Or (instead of `DATABASE_URL`), provide:

- `DB_HOST`
- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_PORT` (optional, default: `5432`)
- `DB_SSLMODE` (optional, default: `require`; `disable` for localhost)

### Optional

- `PORT`: server port (default: `8080`)
- `HTTP_BASE_PATH`: URL prefix (e.g. `/api`, default: empty)
- `CORS_ALLOWED_ORIGINS`: comma-separated list (default: `http://localhost:3000`)

Helm sets typical defaults under `api.env` in [`charts/config-manager/values.yaml`](../charts/config-manager/values.yaml) (e.g. `HTTP_BASE_PATH`, `CORS_ALLOWED_ORIGINS`).

### Authentication (when `auth.enabled=true`)

Set in `application.yaml` under `auth.*` or via `CONFIG_MANAGER_AUTH_*` env overrides. See [auth.md](auth.md).

Secrets (recommended via Kubernetes Secret, not plain values):

- `GOOGLE_CLIENT_ID`: Google OAuth client ID
- `GOOGLE_CLIENT_SECRET`: Google OAuth client secret
- `AUTH_REDIRECT_URL`: OAuth callback URL (Helm can auto-derive from `ingress.host`)
- `AUTH_GOOGLE_HOME_DOMAIN`: optional Google Workspace `hd=` hint
- `AUTH_API_KEYS`: bootstrap API keys as `name:cm_live_...,name:cm_live_...`

Config file keys (under `auth`):

- `auth.enabled`: `true` / `false` (default: `false`)
- `auth.allowedEmailDomains`: list of allowed email domains
- `auth.google.redirectUrl`, `auth.google.homeDomain`
- `auth.session.cookieName`, `cookieSecure`, `idleTimeoutMinutes`, `absoluteTimeoutHours`
- `auth.apiKeys.bootstrapFromEnv`: load `AUTH_API_KEYS` at startup (default: `true`)

## UI (Next.js)

The UI serves `/api/*` via a Route Handler that proxies to the Go API at request time.

### Optional

- `CONFIG_API_BASE_URL`: upstream API base URL for the proxy (default: `http://localhost:8080`). Browser requests to `/api/*` are forwarded here in local dev and when the request hits the UI pod.
- `NEXT_PUBLIC_CONFIG_API_BASE_URL`: fallback when `CONFIG_API_BASE_URL` is unset (server-side only; browser always uses `/api`).
- `CONFIG_API_UPSTREAM_TIMEOUT_MS`: proxy timeout in ms (default: `30000`). Optional; set only if you need a different value.
