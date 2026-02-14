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

- The browser calls the API via same-origin `/api/*`. A Route Handler proxies those requests to `CONFIG_API_BASE_URL` (default: `http://localhost:8080`) at request time.
- `NEXT_PUBLIC_CONFIG_API_BASE_URL` is the server-side fallback when `CONFIG_API_BASE_URL` is unset. See `docs/environment-variables.md` for all UI env vars.
- For full local setup (DB + API + UI), see the root `README.md`.

