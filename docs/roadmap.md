## Done

- [x] **(1) UI cleanup (minimal‑churn, best payoff)** (done 2026-02-01)
  - Added `ui/src/lib/api/hooks.ts` (shared query options + invalidation helpers)
  - Removed duplicated UI types in `ConfigEditor` + `NamespaceBrowserView`
  - Standardized invalidations after create/update/delete

## Next

- [ ] **(2) Backend confidence (small investment, big contributor value)**
  - Add a tiny integration-test baseline: 3–6 high-value tests around create/update/versioning + soft-delete behavior
  - Add CI “check” pipeline: run `make check` on PRs

- [ ] **(3) Packaging / “Airflow-style” distribution (when you’re ready)**
  - Docker images + Helm chart: build/publish `api` + `ui` images and ship a simple chart (service, ingress, env wiring, probes using `/healthz` + `/readyz`)