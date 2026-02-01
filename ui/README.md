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

- Set `NEXT_PUBLIC_CONFIG_API_BASE_URL` to point at the Go API (defaults to `http://localhost:8080`).
- For full local setup (DB + API + UI), see the root `README.md`.

