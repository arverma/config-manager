# Deployment

## Health endpoints

- `GET /healthz`: liveness probe
- `GET /readyz`: readiness probe (checks Postgres reachability)

## Production checklist

- Database
  - Backups configured
  - `DATABASE_URL` uses TLS/SSL in production
- API
  - `CORS_ALLOWED_ORIGINS` restricted to your UI domains
  - Health checks wired to your platform (k8s, ECS, etc.)
  - Logs collected centrally
- UI
  - `NEXT_PUBLIC_CONFIG_API_BASE_URL` points to the deployed API

