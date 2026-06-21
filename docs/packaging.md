# Packaging (Docker + Helm)

This project can be packaged as a single install (similar to Airflow’s chart): UI + API deployed together, with path-based routing.

## Target routing

- `https://example.com/` → UI
- `https://example.com/api/*` → API

## Docker images

**Official public images** (published on release tags to GitHub Container Registry): `ghcr.io/<org>/config-manager-api`, `ghcr.io/<org>/config-manager-ui`. Tags match release versions (e.g. `0.1.0` for tag `v0.1.0`); `latest` points to the latest release.

Build locally:

```bash
docker build -t config-manager-api:dev ./backend
docker build -t config-manager-ui:dev ./ui
```

## Helm chart

Chart source: `charts/config-manager/`

### Authentication (production)

Auth is **disabled by default** (`auth.enabled: false`). For production, enable Google OAuth and API keys via chart values. See [`docs/auth.md`](auth.md) for full setup.

Key values in [`charts/config-manager/values.yaml`](../charts/config-manager/values.yaml):

| Value | Purpose |
|-------|---------|
| `auth.enabled` | Turn on authentication |
| `auth.allowedEmailDomains` | Email domains allowed to sign in |
| `auth.google.existingSecretName` | Secret with `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` |
| `auth.apiKeys.existingSecretName` | Secret with `AUTH_API_KEYS` |
| `ingress.host` | Used to derive OAuth redirect URL when `auth.google.redirectUrl` is empty |

Example with auth:

```bash
helm install config-manager ./charts/config-manager \
  --set ingress.enabled=true \
  --set ingress.host=config.example.com \
  --set ingress.tls.enabled=true \
  --set auth.enabled=true \
  --set auth.allowedEmailDomains={example.com} \
  --set auth.google.existingSecretName=config-manager-oauth \
  --set auth.apiKeys.existingSecretName=config-manager-api-keys \
  --set database.existingSecretName=config-manager-db
```

Parent charts (e.g. [config-manager-ops](https://github.com/arverma/config-manager-ops)) override these under the `config-manager:` subchart key.

### Redis caching (production, optional)

Caching is **disabled by default** (`cache.enabled: false`). For multi-replica API deployments, enable external Redis. See [`docs/caching.md`](caching.md).

| Value | Purpose |
|-------|---------|
| `cache.enabled` | Turn on Redis caching and Redis-backed sessions |
| `cache.redis.host` | External Redis hostname or IP |
| `cache.redis.port` | Redis port (default `6379`) |
| `cache.redis.existingSecretName` | Secret with `REDIS_PASSWORD` (or `REDIS_URL`) |
| `cache.config.ttlSeconds` | Safety TTL for cached GET latest responses |

No Bitnami Redis subchart is bundled — use your own managed or self-hosted Redis.

## Install from the public Helm repo (GitHub Pages)

Once published, the chart can be installed via `helm repo add`:

```bash
helm repo add config-manager https://<org>.github.io/<repo>
helm repo update

helm install config-manager config-manager/config-manager --version <chart-version> \
  --set ingress.enabled=true \
  --set ingress.host=example.com
```

Current chart version in this repo: `0.1.2` (see `charts/config-manager/Chart.yaml`).

Install with an external Postgres `DATABASE_URL` secret (simplest):

```bash
kubectl create secret generic config-manager-db \
  --from-literal=DATABASE_URL='postgres://USER:PASS@HOST:5432/DB?sslmode=require'

helm install config-manager ./charts/config-manager \
  --set database.existingSecretName=config-manager-db \
  --set ingress.enabled=true \
  --set ingress.host=example.com
```

With External Secrets Operator (ESO) + Vault: enable `api.externalSecret` to sync a Vault path into a Secret; the API uses `envFrom`. Store `DB_PASSWORD` (and other secrets) in Vault; keys must be valid env var names.

```bash
helm install config-manager ./charts/config-manager \
  --set api.externalSecret.enabled=true \
  --set api.externalSecret.clusterSecretStore.name='cluster_secret_store_name' \
  --set api.externalSecret.vaultPath='path/to/secret' \
  --set database.parts.host='YOUR_DB_HOST' \
  --set database.parts.name='YOUR_DB_NAME' \
  --set database.parts.user='YOUR_DB_USER' \
  --set ingress.enabled=true \
  --set ingress.host=example.com
```

In this mode, store the Postgres password in Vault as `DB_PASSWORD` (so it becomes an env var), and the API will assemble the connection string from `DB_HOST/DB_PORT/DB_NAME/DB_USER/DB_PASSWORD` (optionally `DB_SSLMODE`).

## Publishing (release tag)

On push of a tag `v*` (e.g. `v0.1.4`), GitHub Actions (`.github/workflows/release-chart.yaml`):

1. **Helm chart**: Publishes the chart to the `gh-pages` branch (Helm repo at `https://<org>.github.io/<repo>/index.yaml`).
2. **Docker images**: Builds and pushes API and UI images to GHCR (`ghcr.io/<org>/config-manager-api`, `ghcr.io/<org>/config-manager-ui`) with the same version tag (e.g. `0.1.0`) and `latest`.

## Publishing to GCP Artifact Registry (GAR)

### 1) Create a GAR repo (one-time)

```bash
gcloud artifacts repositories create REPO \
  --repository-format=docker \
  --location=LOCATION
```

### 2) Configure Docker auth

```bash
gcloud auth configure-docker LOCATION-docker.pkg.dev
```

### 3) Push images

```bash
export LOCATION=LOCATION
export PROJECT=PROJECT
export REPO=REPO

docker tag config-manager-api:dev "$LOCATION-docker.pkg.dev/$PROJECT/$REPO/config-manager-api:dev"
docker tag config-manager-ui:dev "$LOCATION-docker.pkg.dev/$PROJECT/$REPO/config-manager-ui:dev"

docker push "$LOCATION-docker.pkg.dev/$PROJECT/$REPO/config-manager-api:dev"
docker push "$LOCATION-docker.pkg.dev/$PROJECT/$REPO/config-manager-ui:dev"
```

### 4) Push the Helm chart as OCI

Helm 3 supports pushing charts to OCI registries:

```bash
export CHART_VERSION=0.1.2

helm package ./charts/config-manager --version "$CHART_VERSION"
helm push "config-manager-$CHART_VERSION.tgz" "oci://$LOCATION-docker.pkg.dev/$PROJECT/$REPO/charts"
```

Install from OCI:

```bash
helm install config-manager "oci://$LOCATION-docker.pkg.dev/$PROJECT/$REPO/charts/config-manager" \
  --version "$CHART_VERSION" \
  --set database.existingSecretName=config-manager-db \
  --set ingress.enabled=true \
  --set ingress.host=example.com
```

