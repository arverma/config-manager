# Environment variables

## Backend (API)

### Required

- `DATABASE_URL`: Postgres connection string

### Optional

- `PORT`: server port (default: `8080`)
- `CORS_ALLOWED_ORIGINS`: comma-separated list of allowed origins (default: `http://localhost:3000`)

## UI (Next.js)

### Recommended

- `NEXT_PUBLIC_CONFIG_API_BASE_URL`: API base URL for the browser (e.g. `http://localhost:8080`)

### Optional

- `CONFIG_API_BASE_URL`: server-side fallback base URL (used when `NEXT_PUBLIC_...` is not set)

