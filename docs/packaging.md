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

## Install from the public Helm repo (GitHub Pages)

Once published, the chart can be installed via `helm repo add`:

```bash
helm repo add config-manager https://<org>.github.io/<repo>
helm repo update

helm install config-manager config-manager/config-manager --version 0.1.0 \
  --set ingress.enabled=true \
  --set ingress.host=example.com
```

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

On push of a tag `v*` (e.g. `v0.1.0`), GitHub Actions:

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
export CHART_VERSION=0.1.0

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

