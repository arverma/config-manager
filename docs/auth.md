# Authentication

Config Manager supports **Google OAuth** for browser users and **API keys** for machine clients (SDKs, CI/CD). Authentication is **disabled by default** for local development.

When `auth.enabled=true`, all API routes except health checks and OAuth login/callback require a valid session cookie or API key.

## v1 behavior

- Any authenticated Google user (from an allowed email domain) can read and write.
- Role-based access (viewer vs developer) and per-namespace ACL are planned for a future release. See [rbac.md](rbac.md).

## Google OAuth setup

### 1. Create OAuth credentials

In [Google Cloud Console](https://console.cloud.google.com/apis/credentials):

1. Create an **OAuth 2.0 Client ID** (Web application).
2. Add authorized redirect URI:
   - Production: `https://<your-host>/api/auth/callback/google`
   - Local auth testing (via UI proxy): `http://localhost:3000/api/auth/callback/google`

### 2. Configure allowed email domains

Only users whose email domain matches `auth.allowedEmailDomains` can sign in.

Example Helm values:

```yaml
auth:
  enabled: true
  allowedEmailDomains:
    - example.com
    - partner.org
  google:
    existingSecretName: config-manager-oauth
    homeDomain: example.com   # optional Google Workspace hd= hint
  session:
    cookieSecure: true
  apiKeys:
    existingSecretName: config-manager-api-keys
```

### 3. Create Kubernetes secrets (no plaintext in values)

**OAuth secret** (`config-manager-oauth`):

```bash
kubectl create secret generic config-manager-oauth \
  --from-literal=GOOGLE_CLIENT_ID='...apps.googleusercontent.com' \
  --from-literal=GOOGLE_CLIENT_SECRET='...'
```

**API keys secret** (`config-manager-api-keys`):

```bash
kubectl create secret generic config-manager-api-keys \
  --from-literal=AUTH_API_KEYS='ci:cm_live_...,etl:cm_live_...'
```

Generate keys with the CLI (see below) before storing them in the secret.

### 4. Install with auth enabled

```bash
helm install config-manager ./charts/config-manager \
  --set ingress.enabled=true \
  --set ingress.host=config.example.com \
  --set ingress.tls.enabled=true \
  --set ingress.tls.secretName=config-tls \
  --set auth.enabled=true \
  --set auth.allowedEmailDomains={example.com} \
  --set auth.google.existingSecretName=config-manager-oauth \
  --set auth.apiKeys.existingSecretName=config-manager-api-keys
```

The chart auto-derives `AUTH_REDIRECT_URL` from `ingress.host` when `auth.google.redirectUrl` is empty.

## Browser login flow

1. User opens the UI → `GET /api/auth/session`
2. If unauthenticated → redirect to `GET /api/auth/login/google?returnTo=/configs`
3. Google callback → session cookie (`cm_session`) set on the shared domain
4. Subsequent `fetch("/api/...")` calls include the cookie automatically

Logout (`POST /api/auth/logout`) clears the Config Manager session only; it does **not** sign the user out of Google.

## API keys (machine clients)

### Create a key

```bash
cd backend
DATABASE_URL='postgres://...' go run ./cmd/config-manager auth create-api-key --name ci-pipeline
```

Prints `cm_live_<secret>` once. Store it securely.

### Use from SDK / curl

```bash
curl -H "Authorization: Bearer cm_live_..." https://config.example.com/api/namespaces
```

API keys are stored hashed in Postgres. Keys can also be bootstrapped from the `AUTH_API_KEYS` env var at startup (`name:key,name:key`).

## `created_by` audit field

When authentication is enabled, `created_by` on config versions is **set by the server** from the authenticated identity (user email or `apikey:<name>`). Clients must not send `created_by`; doing so returns `400 client_created_by_not_allowed`.

## Session storage

By default (`cache.enabled=false`), browser sessions are stored in Postgres (`auth_sessions`).

When [Redis caching](caching.md) is enabled (`cache.enabled=true`), sessions and OAuth state are stored in Redis instead. Existing Postgres sessions are not migrated — enabling cache logs users out once.

## Local development

Keep `auth.enabled: false` (default) for the standard two-terminal workflow (`make api-run`, `make ui-dev`).

To test auth locally, enable auth in `backend/confs/application.yaml`, register the localhost redirect URI in Google Cloud Console, and set `auth.session.cookieSecure: false`.

## Configuration reference

See [environment-variables.md](environment-variables.md) for env vars and [charts/config-manager/values.yaml](../charts/config-manager/values.yaml) for Helm options.
