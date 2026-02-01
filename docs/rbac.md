# RBAC (Future)

This project will eventually support **two user experiences**:

- **Viewer**: browse and read configs (and versions) only
- **Developer**: create namespaces/configs and write new versions

## Why this matters now

Even before auth is added, we keep the API/UI structure compatible with RBAC by:

- Keeping **read** operations cleanly separated from **write** operations
- Making the UI a thin client over the REST API, so feature gating is straightforward

## Proposed permissions model

### Viewer (read-only)

- `GET /namespaces`
- `GET /namespaces/{namespace}/browse?prefix=...`
- `GET /configs/{namespace}/{path}`
- `GET /configs/{namespace}/{path}/versions`
- `GET /configs/{namespace}/{path}/versions/{version}`

### Developer (write)

All viewer permissions, plus:

- `POST /namespaces`
- `POST /configs/{namespace}/{path}`
- `PUT /configs/{namespace}/{path}`
- `DELETE /configs/{namespace}/{path}/versions/{version}` (non-latest only)

## Authn/Authz approach (later)

When we introduce authentication (JWT/OIDC or API keys), we can add:

- A middleware in the Go API that resolves an **actor** (user/service)
- An authorization layer that maps actor → role(s) → endpoint access
- Optional scoping rules:
  - per namespace
  - per path prefix

The OpenAPI contract should remain stable; only auth headers and error responses (401/403) will be added.

