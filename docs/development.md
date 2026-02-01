# Development

## Two-terminal workflow

Terminal 1:

```bash
make db-up
make db-apply
make api-run
```

Terminal 2:

```bash
make ui-install
make ui-dev
```

## Common tasks

- `make check`: backend tests + UI lint/typecheck
- `make smoke`: quick API smoke test
- `make db-reset`: reset DB volume

## Postman collection (from OpenAPI)

You can generate a Postman collection instantly by importing our OpenAPI spec:

- In Postman: **Import** → **File** → select `api/openapi.yaml`
- Set the base URL to your local API (typically `http://localhost:8080`) or use the default.

