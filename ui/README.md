# UI (Next.js)

Minimal UI scaffolded to mirror the REST URL shape:

`/configs/{namespace}/{path}`

## Run locally

```bash
npm install
cp .env.example .env.local
npm run dev
```

## Notes

- In the browser, the UI calls the API via same-origin `/api/*`.
- In local dev, Next.js rewrites `/api/*` to `CONFIG_API_BASE_URL` (default: `http://localhost:8080`).
- `NEXT_PUBLIC_CONFIG_API_BASE_URL` is only used as a server-side fallback when `CONFIG_API_BASE_URL` is not set.
- For full local setup (DB + API + UI), see the root `README.md`.

