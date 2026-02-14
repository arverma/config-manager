package migrations

import "embed"

// FS holds the migration SQL files. The API runs these on startup via golang-migrate.
//go:embed *.sql
var FS embed.FS
