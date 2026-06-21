# Backend (Go API)

This service exposes the REST API defined in `../api/openapi.yaml` (source of truth). Config is loaded from `application.yaml` (default `confs/application.yaml` or `-config` path) and overridden by `CONFIG_MANAGER_*` env vars.

For full local setup (DB + API + UI), see the root `README.md`.

## Run locally (via make)

```bash
make db-up
make api-run
```

The API runs DB migrations on startup (from `backend/migrations/`). Authentication and Redis caching are disabled by default; see [`../docs/auth.md`](../docs/auth.md) and [`../docs/caching.md`](../docs/caching.md).

## Health

- `GET /healthz`: liveness
- `GET /readyz`: readiness (checks Postgres reachability)

When `HTTP_BASE_PATH=/api`, prefix paths with `/api` (e.g. `/api/healthz`).

## Authentication routes

When `auth.enabled=true`:

| Route | Purpose |
|-------|---------|
| `GET /auth/login/google` | Redirect to Google OAuth (`?returnTo=` optional) |
| `GET /auth/callback/google` | OAuth callback; sets session cookie |
| `GET /auth/session` | Current session (or `{ auth_enabled: false }` when auth off) |
| `POST /auth/logout` | Clear application session |

Protected routes accept either a session cookie or `Authorization: Bearer cm_live_...`.

## CLI

Create an API key (requires database access):

```bash
DATABASE_URL='postgres://...' go run ./cmd/config-manager auth create-api-key --name ci-pipeline
```

Prints `cm_live_<secret>` once.

## Migrations

Schema migrations live in `backend/migrations/`. Notable:

- `000003_auth` — `auth_sessions`, `auth_oauth_states`, `api_keys`
