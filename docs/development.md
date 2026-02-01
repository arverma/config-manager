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
make ui-dev
```

## Common tasks

- `make check`: backend tests + UI lint/typecheck
- `make smoke`: quick API smoke test
- `make db-reset`: reset DB volume

