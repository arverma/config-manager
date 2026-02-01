# Backend (Go API)

This service exposes the REST API defined in `../api/openapi.yaml` (source of truth).

For full local setup (DB + API + UI), see the root `README.md`.

## Run locally (via make)

```bash
make db-up
make db-apply
make api-run
```

## Health

- `GET /healthz`: liveness
- `GET /readyz`: readiness (checks Postgres reachability)

