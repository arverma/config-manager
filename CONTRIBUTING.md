# Contributing

## Prerequisites

- Docker + Docker Compose (Postgres)
- Go (API)
- Node.js + npm (UI)

## Local development

Terminal 1:

```bash
make db-up
make api-run
```

(The API runs DB migrations on startup.)

Terminal 2:

```bash
make ui-install
make ui-dev
```

## Before opening a PR

```bash
make check
make smoke
```

## Testing roadmap (backend)

The Go backend currently has minimal automated coverage. High‑ROI tests to add next:

- **Pagination correctness**
  - `GET /configs?recursive=false` cursor correctness (no skips/duplicates)
  - `GET /namespaces/{namespace}/browse` cursor correctness
- **Soft delete semantics (configs)**
  - After `DELETE /configs/{namespace}/{path}`, config is hidden from list/browse and `GET /configs/{namespace}/{path}` returns 404
- **Version rules**
  - `DELETE /configs/{namespace}/{path}/versions/{version}` cannot delete latest
  - `PUT /configs/{namespace}/{path}` returns 409 with `code=no_change` when body is unchanged

See `docs/testing.md` for the recommended test harness approach.

## Project structure

- `api/openapi.yaml`: API contract (update when behavior changes)
- `backend/`: Go API
- `ui/`: Next.js UI
- `db/`: legacy schema reference; source of truth is `backend/migrations/` (API runs these on startup)
- `docs/`: architecture and operational docs

## Conventions

- **API changes**: update `api/openapi.yaml` in the same PR.
- **DB changes**: add versioned migrations under `backend/migrations/` (e.g. `000002_description.up.sql` and `.down.sql`); document in the PR.
- **Go**: keep handlers small; prefer shared helpers over copy/paste.
- **UI**: prefer shared hooks/utilities over per-component fetch logic.

