# RBAC

## v1 (current): authentication only

When `auth.enabled=true`:

- **Browser users** sign in with Google OAuth (allowed email domains only).
- **Machine clients** use API keys (`Authorization: Bearer cm_live_...`).
- Any authenticated principal can read and write all namespaces and configs.

See [auth.md](auth.md) for setup.

## v2 (planned): roles and scoped access

Future releases will add:

- **Viewer**: browse and read configs (and versions) only
- **Developer**: create/update/delete namespaces and configs (within API rules)
- Optional scoping: per namespace, per path prefix

### Proposed permissions model

#### Viewer (read-only)

- `GET /namespaces`
- `GET /namespaces/{namespace}/browse?prefix=...`
- `GET /configs/{namespace}/{path}`
- `GET /configs/{namespace}/{path}/versions`
- `GET /configs/{namespace}/{path}/versions/{version}`

#### Developer (write)

All viewer permissions, plus:

- `POST /namespaces`
- `POST /configs/{namespace}/{path}`
- `PUT /configs/{namespace}/{path}`
- `DELETE /configs/{namespace}/{path}` (hard delete)
- `DELETE /configs/{namespace}/{path}/versions/{version}` (non-latest only)
- `DELETE /namespaces/{namespace}` (allowed only when empty)

## Design notes

The API/UI structure keeps read and write operations separated so role gating can be added without changing the URL model. Authorization will map actor → role(s) → endpoint access in Go middleware.

The OpenAPI contract should remain stable; auth uses session cookies and bearer API keys with standard `401`/`403` responses.
