# Packaging (Docker + Helm)

This project can be packaged as a single install (similar to Airflow’s chart): UI + API deployed together, with path-based routing.

## Target routing

- `https://example.com/` → UI
- `https://example.com/api/*` → API

## Docker images

Build locally:

```bash
docker build -t config-manager-api:dev ./backend
docker build -t config-manager-ui:dev ./ui
```

## Helm chart

Chart source: `charts/config-manager/`

Install with an external Postgres `DATABASE_URL` secret (simplest):

```bash
kubectl create secret generic config-manager-db \
  --from-literal=DATABASE_URL='postgres://USER:PASS@HOST:5432/DB?sslmode=require'

helm install config-manager ./charts/config-manager \
  --set database.existingSecretName=config-manager-db \
  --set ingress.enabled=true \
  --set ingress.host=example.com
```

Or, if you use External Secrets Operator (ESO) + Vault, have it extract **all** key/value pairs from a Vault path and expose them to the API pod (via `envFrom`) (recommended when you want password-only secrets and room to add more env vars later).

This is ideal when you want to keep adding keys at the same Vault path over time and have them available as environment variables (keys must be valid env var names).

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

