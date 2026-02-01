# Backend (Go API)

This service exposes the REST API defined in `../api/openapi.yaml`.

For full local setup (DB + API + UI), see the root `README.md`.

## Run locally (API only)

1. Start Postgres:

```bash
docker compose up -d postgres
```

2. Apply schema:

```bash
PGPASSWORD=postgres psql -h 127.0.0.1 -p 5432 -U postgres -d config_manager -f "../db/migrations/001_init.sql"
```

3. Run the API:

 - macOS/Linux (bash/zsh):

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/config_manager?sslmode=disable"
go run ./cmd/config-manager
```

 - Windows PowerShell:

```powershell
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5432/config_manager?sslmode=disable"
go run ./cmd/config-manager
```

## Contract

See `../api/openapi.yaml`.

