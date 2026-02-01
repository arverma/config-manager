# Config Manager

Postgres-backed config registry with immutable versioning; **latest = highest version number**.

## Repo layout

- `api/openapi.yaml` - REST API contract (OpenAPI-first; used for future SDK generation)
- `db/migrations/` - Postgres migrations
- `docs/architecture.md` - architecture + versioning flows (mermaid diagrams)
- `backend/` - Go API service (Chi router)
- `ui/` - Next.js UI (self-serve config management)

## Identity model

A logical config is uniquely identified by:

`(namespace, path)`

And addressed as:

`/configs/{namespace}/{path}`

Where `format` (`json|yaml`) is an attribute set at create-time (one format per config identity).

## Local setup (OS-independent)

### Prerequisites

- **Docker** (for Postgres) + Docker Compose
- **Go** (for the API)
- **Node.js + npm** (for the UI)

Optional:

- `psql` (if you prefer applying schema from your host; otherwise use Docker-based apply below)

### Quickstart (recommended)

If you have `make`:

```bash
make db-up
make db-apply
make api-run
```

In a second terminal:

```bash
make ui-dev
```

Open the UI at `http://localhost:3000` (API is `http://localhost:8080`).

### Manual setup (no `make`)

1. Start Postgres:

```bash
docker compose up -d postgres
```

2. Apply schema:

- **Option A (host `psql`)**:

```bash
PGPASSWORD=postgres psql -h 127.0.0.1 -p 5432 -U postgres -d config_manager -f "db/migrations/001_init.sql"
```

- **Option B (Docker-only, no host `psql`)**:

```bash
docker compose cp "db/migrations/001_init.sql" postgres:/tmp/001_init.sql
docker compose exec -T postgres psql -U postgres -d config_manager -f /tmp/001_init.sql
```

3. Run the API (default `:8080`):

- macOS/Linux (bash/zsh):

```bash
cd backend
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/config_manager?sslmode=disable"
go run ./cmd/config-manager
```

- Windows PowerShell:

```powershell
cd backend
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5432/config_manager?sslmode=disable"
go run ./cmd/config-manager
```

4. Run the UI (default `:3000`):

```bash
cd ui
npm install
cp .env.example .env.local
npm run dev
```

### Environment variables

- **API**
  - `DATABASE_URL` (required)
  - `PORT` (optional; defaults to `8080`)
- **UI**
  - `NEXT_PUBLIC_CONFIG_API_BASE_URL` (recommended)
  - `CONFIG_API_BASE_URL` (server-side only fallback)

### Handy commands

See `Makefile`:

```bash
make help
make smoke
make check
make db-reset
```


