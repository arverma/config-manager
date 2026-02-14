# Backend (Go API)

This service exposes the REST API defined in `../api/openapi.yaml` (source of truth). Config is loaded from `application.yaml` (default `confs/application.yaml` or `-config` path) and overridden by `CONFIG_MANAGER_*` env vars.

For full local setup (DB + API + UI), see the root `README.md`.

## Run locally (via make)

```bash
make db-up
make api-run
```

The API runs DB migrations on startup (from `backend/migrations/`).

## Health

- `GET /healthz`: liveness
- `GET /readyz`: readiness (checks Postgres reachability)

