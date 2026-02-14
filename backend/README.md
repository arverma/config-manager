# Backend (Go API)

This service exposes the REST API defined in `../api/openapi.yaml` (source of truth).

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

