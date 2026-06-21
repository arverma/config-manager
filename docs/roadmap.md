## Done

- [x] **(1) UI cleanup (minimal‑churn, best payoff)** (done 2026-02-01)
  - Added `ui/src/lib/api/hooks.ts` (shared query options + invalidation helpers)
  - Removed duplicated UI types in `ConfigEditor` + `NamespaceBrowserView`
  - Standardized invalidations after create/update/delete

- [x] **(4) Authentication v1** (done 2026-06)
  - Google OAuth in the Go API with domain allowlist
  - Postgres-backed sessions (sliding idle timeout) and API keys (`cm_live_...`)
  - Server-side `created_by`; Helm `auth.*` values and secrets; UI auth gate
  - See [`docs/auth.md`](auth.md)

- [x] **(7) Optional Redis caching** (done 2026-06)
  - Disabled by default; external Redis when `cache.enabled=true` (no Bitnami subchart)
  - GET latest config cache-aside with write invalidation; sessions + OAuth state in Redis when enabled
  - Helm `cache.*` values; credentials via Secret / ESO
  - See [`docs/caching.md`](caching.md)

## Partially done

- [ ] **(2) Backend confidence**
  - Auth tests landed (`backend/internal/auth/*_test.go`, `backend/internal/httpapi/auth_integration_test.go`)
  - Still missing: pagination, hard-delete, and version-rule integration tests
  - Still missing: CI “check” pipeline (`make check` on PRs)

## Next

- [ ] **(2) Backend confidence** (finish)
  - Add high-value integration tests: pagination cursors, hard deletes, version rules (`no_change`, cannot delete latest)
  - Add GitHub Actions workflow to run `make check` on pull requests

- [ ] **(3) Packaging polish**
  - Chart + GHCR images already publish on `v*` tags
  - Document auth in Helm install examples; refine parent-chart override patterns (see [config-manager-ops](https://github.com/arverma/config-manager-ops))

- [ ] **(5) RBAC v2**
  - Viewer vs developer roles; optional per-namespace / per-path ACL
  - See [`docs/rbac.md`](rbac.md)

- [ ] **(6) SDK expansion**
  - Java SDK shipped separately: [config-manager-java-sdk](https://github.com/arverma/config-manager-java-sdk)
  - Go/Python clients from `api/openapi.yaml` remain future work
