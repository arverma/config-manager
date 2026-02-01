# Environment variables

## Backend (API)

### Required

- `DATABASE_URL`: Postgres connection string

Or (instead of `DATABASE_URL`), provide the parts:

- `DB_HOST`
- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_PORT` (optional, default: `5432`)
- `DB_SSLMODE` (optional, default: `require`, or `disable` for `localhost`)

### Optional

- `PORT`: server port (default: `8080`)
- `HTTP_BASE_PATH`: mount the API under a URL prefix (e.g. `/api`, default: empty)
- `CORS_ALLOWED_ORIGINS`: comma-separated list of allowed origins (default: `http://localhost:3000`)

## UI (Next.js)

### Optional

- `CONFIG_API_BASE_URL`: upstream API base URL used by Next.js to proxy `/api/*` via rewrites (default: `http://localhost:8080`). This is how browser requests to `/api/*` reach the Go API in local dev.
- `NEXT_PUBLIC_CONFIG_API_BASE_URL`: fallback API base URL for **server-side** requests when `CONFIG_API_BASE_URL` is unset (not used by the browser; the browser always calls `/api`).

