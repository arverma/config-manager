# Testing

This project is early-stage and intentionally lightweight.

## Current state

- `make check` runs:
  - Go: `go test ./...` (currently no test files)
  - UI: lint + typecheck
- `make smoke` runs a small Node-based API smoke flow against a running API.

## Backend test harness (recommended approach)

Contributors can add an `httptest` suite for the Go API by running it against a real Postgres instance.

### Option A (minimal dependencies): Docker-based Postgres

Requirements:
- Docker running locally

Strategy:
- Start Postgres (docker compose)
- Create a temporary database per test run (or per package)
- Run migrations (the API runs them on startup, or run the same migrations from `backend/migrations/` in test setup)
- Create a `pgxpool.Pool` to that DB
- Use `httptest.NewServer(httpapi.NewRouter(pool))` to run request/response tests

Suggested setup steps in tests:

1. Start Postgres (outside the test process):

```bash
make db-up
```

2. In tests:
   - connect to `postgres` database using admin creds
   - `CREATE DATABASE config_manager_test_<random>`
   - connect to that DB and apply migrations (e.g. run the API against it briefly, or run migration SQL from `backend/migrations/`)
   - run tests
   - `DROP DATABASE ...` in cleanup

This keeps the codebase dependency-light and works well in CI using a Postgres service.

### Option B (hermetic): testcontainers-go

If we want the tests to be fully hermetic (no pre-running services), use `testcontainers-go`
to start Postgres from Go tests. This adds a dependency but simplifies developer setup.

## What to test first (high ROI)

- **Pagination correctness**
  - `GET /configs?recursive=false` cursor correctness (no skips/duplicates)
  - `GET /namespaces/{namespace}/browse` cursor correctness
- **Soft delete semantics (configs)**
  - After `DELETE /configs/{namespace}/{path}`, config is hidden from list/browse and `GET /configs/{namespace}/{path}` returns 404
- **Version rules**
  - Cannot delete latest version
  - `PUT` returns 409 with `code=no_change` when body is unchanged

