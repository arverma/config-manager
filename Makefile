.DEFAULT_GOAL := help

# Config Manager Makefile
#
# Usage:
#   make <target>
#
# Notes:
# - DB runs via docker compose (see docker-compose.yaml)
# - API reads DATABASE_URL
# - UI reads NEXT_PUBLIC_CONFIG_API_BASE_URL / CONFIG_API_BASE_URL

PROJECT := config-manager

DB_HOST ?= 127.0.0.1
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASS ?= postgres
DB_NAME ?= config_manager

DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASS)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable
API_PORT ?= 8080

.PHONY: help
help:
	@printf "\nTargets:\n"
	@printf "  dev               Print local dev workflow\n"
	@printf "  db-up             Start Postgres (docker compose)\n"
	@printf "  db-down           Stop Postgres\n"
	@printf "  db-reset          Stop Postgres + delete volume\n"
	@printf "  db-psql           Open psql shell\n"
	@printf "  db-apply          Apply db/schema.sql\n"
	@printf "  db-drop-schema    Drop + recreate public schema (destructive)\n"
	@printf "\n"
	@printf "  api-run           Run Go API (PORT=$(API_PORT))\n"
	@printf "  api-test          Run Go tests\n"
	@printf "  api-fmt           gofmt backend files\n"
	@printf "  api-build         Build backend binary (bin/)\n"
	@printf "\n"
	@printf "  ui-install        npm install (ui/)\n"
	@printf "  ui-dev            Run Next dev server (ui/)\n"
	@printf "  ui-lint           eslint (ui/)\n"
	@printf "  ui-typecheck      TypeScript check (ui/)\n"
	@printf "  ui-check          Lint + typecheck (ui/)\n"
	@printf "  ui-build          Build UI (Next)\n"
	@printf "\n"
	@printf "  fmt               Format backend code\n"
	@printf "  lint              Lint/typecheck (ui) + vet (api)\n"
	@printf "  test              Run backend tests\n"
	@printf "  check             api-test + ui-check\n"
	@printf "  smoke             Quick API smoke (namespaces/configs/versions)\n"
	@printf "\n"

.PHONY: dev
dev:
	@printf "\nLocal dev:\n\n"
	@printf "Terminal 1:\n"
	@printf "  make db-up\n"
	@printf "  make db-apply\n"
	@printf "  make api-run\n\n"
	@printf "Terminal 2:\n"
	@printf "  make ui-dev\n\n"

.PHONY: db-up
db-up:
	docker compose up -d postgres

.PHONY: db-down
db-down:
	docker compose down

.PHONY: db-reset
db-reset:
	docker compose down -v
	docker compose up -d postgres

.PHONY: db-psql
db-psql:
	PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME)

.PHONY: db-apply
db-apply:
	PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f "db/schema.sql"

.PHONY: db-drop-schema
db-drop-schema:
	PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'

.PHONY: api-run
api-run:
	cd backend && PORT=$(API_PORT) DATABASE_URL="$(DATABASE_URL)" go run ./cmd/config-manager

.PHONY: api-test
api-test:
	cd backend && go test ./...

.PHONY: api-fmt
api-fmt:
	cd backend && gofmt -w ./...

.PHONY: api-build
api-build:
	@mkdir -p bin
	cd backend && go build -o ../bin/config-manager ./cmd/config-manager

.PHONY: ui-install
ui-install:
	cd ui && npm install

.PHONY: ui-dev
ui-dev:
	cd ui && cp .env.example .env.local && npm run dev

.PHONY: ui-lint
ui-lint:
	cd ui && npm run lint

.PHONY: ui-typecheck
ui-typecheck:
	cd ui && npx tsc -p tsconfig.json --noEmit

.PHONY: ui-check
ui-check:
	cd ui && npm run lint && npx tsc -p tsconfig.json --noEmit

.PHONY: ui-build
ui-build:
	cd ui && npm run build

.PHONY: fmt
fmt: api-fmt

.PHONY: lint
lint:
	cd backend && go vet ./...
	$(MAKE) ui-check

.PHONY: test
test: api-test

.PHONY: check
check: api-test ui-check

.PHONY: smoke
smoke:
	@node - <<-'NODE'
	const base = `http://127.0.0.1:${process.env.API_PORT || "$(API_PORT)"}`;
	async function req(path, opts) {
	  const r = await fetch(base + path, opts);
	  const t = await r.text();
	  return { status: r.status, text: t };
	}
	(async () => {
	  console.log("base", base);
	  console.log("create_ns", (await req("/namespaces", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ name: "platform" }) })).status);
	  console.log("list_ns", (await req("/namespaces", {})).status);
	  console.log("create_cfg", (await req("/configs/platform/collector", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ format: "yaml", body_raw: "key: v1\n" }) })).status);
	  console.log("put_v2", (await req("/configs/platform/collector", { method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ body_raw: "key: v2\n" }) })).status);
	  const v = await req("/configs/platform/collector/versions", {});
	  console.log("versions", v.status, v.text.slice(0, 120));
	})().catch((e) => {
	  console.error(e);
	  process.exit(1);
	});
	NODE

