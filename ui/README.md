# UI (Next.js)

Minimal UI scaffolded to mirror the REST URL shape:

`/configs/{namespace}/{path}`

## Run locally

From repo root, `make ui-dev` runs `npm run dev` and copies `.env.example` → `.env.local`. Or from `ui/`:

```bash
npm install
cp .env.example .env.local
npm run dev
```

## Notes

- The browser calls the API via same-origin `/api/*`. A Route Handler proxies those requests to `CONFIG_API_BASE_URL` (default: `http://localhost:8080`) at request time.
- `NEXT_PUBLIC_CONFIG_API_BASE_URL` is the server-side fallback when `CONFIG_API_BASE_URL` is unset. See `docs/environment-variables.md` for all UI env vars.
- For full local setup (DB + API + UI), see the root `README.md`.

## Authentication

When the API has `auth.enabled=true`:

- `AuthGate` (in `src/components/AuthGate.tsx`) checks `GET /api/auth/session` on load.
- Unauthenticated users are redirected to `/api/auth/login/google`.
- `ConfigsHeader` shows the signed-in email and a Logout button.

Local dev uses auth **off** by default (`auth.enabled: false` in `backend/confs/application.yaml`), so no sign-in is required. In production (same-origin `/api/*` via ingress), session cookies are sent automatically on API calls.

See [`../docs/auth.md`](../docs/auth.md) for enabling auth locally or in Helm.
