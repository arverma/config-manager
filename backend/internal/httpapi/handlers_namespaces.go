package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var namespaceNameRE = regexp.MustCompile(`^[a-z_]+$`)

func handleListNamespaces(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool) {
	limit, err := parseLimit(req, 50)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}
	offset, err := parseCursorOffset(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}

	rows, err := db.Query(req.Context(), `
		SELECT
			n.id, n.name, n.created_at, n.updated_at,
			COUNT(c.id) AS config_count
		FROM namespaces n
		LEFT JOIN configs c
			ON c.namespace = n.name
		GROUP BY n.id, n.name, n.created_at, n.updated_at
		ORDER BY n.name ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}
	defer rows.Close()

	items := make([]NamespaceWithCount, 0, limit)
	for rows.Next() {
		var id pgtype.UUID
		var n NamespaceWithCount
		var count int64
		if err := rows.Scan(&id, &n.Name, &n.CreatedAt, &n.UpdatedAt, &count); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "scan failed", nil)
			return
		}
		n.ID = uuidToString(id)
		n.ConfigCount = int(count)
		items = append(items, n)
	}

	var next *string
	if len(items) == limit {
		c := encodeCursorOffset(offset + limit)
		next = &c
	}

	writeJSON(w, http.StatusOK, NamespaceListResponse{
		Items:      items,
		NextCursor: next,
	})
}

func handleCreateNamespace(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool, name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "name is required", map[string]any{"field": "name"})
		return
	}
	if !namespaceNameRE.MatchString(name) {
		writeError(w, http.StatusBadRequest, "bad_request", "name must match ^[a-z_]+$", map[string]any{"field": "name"})
		return
	}

	var id pgtype.UUID
	var createdAt, updatedAt pgtype.Timestamptz

	err := db.QueryRow(req.Context(), `
		INSERT INTO namespaces (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`, name).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "conflict", "namespace already exists", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "insert failed", nil)
		return
	}

	writeJSON(w, http.StatusCreated, Namespace{
		ID:        uuidToString(id),
		Name:      name,
		CreatedAt: createdAt.Time,
		UpdatedAt: updatedAt.Time,
	})
}

func handleDeleteNamespace(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool, namespace string) {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "namespace is required", nil)
		return
	}
	if !namespaceNameRE.MatchString(namespace) {
		writeError(w, http.StatusBadRequest, "bad_request", "namespace must match ^[a-z_]+$", nil)
		return
	}

	tx, err := db.Begin(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "begin failed", nil)
		return
	}
	defer tx.Rollback(req.Context())

	// Ensure namespace exists and lock it.
	var exists bool
	if err := tx.QueryRow(req.Context(), `
		SELECT EXISTS(SELECT 1 FROM namespaces WHERE name = $1)
	`, namespace).Scan(&exists); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "not_found", "namespace not found", nil)
		return
	}

	// Count configs in namespace.
	var cnt int64
	if err := tx.QueryRow(req.Context(), `
		SELECT COUNT(*) FROM configs WHERE namespace = $1
	`, namespace).Scan(&cnt); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}
	if cnt > 0 {
		writeError(w, http.StatusConflict, "conflict", "namespace is not empty", map[string]any{"config_count": cnt})
		return
	}

	// Delete namespace. (If a race inserts a config, FK will block; treat as conflict.)
	cmd, err := tx.Exec(req.Context(), `DELETE FROM namespaces WHERE name = $1`, namespace)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			writeError(w, http.StatusConflict, "conflict", "namespace is not empty", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "delete failed", nil)
		return
	}
	if cmd.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "not_found", "namespace not found", nil)
		return
	}

	if err := tx.Commit(req.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "commit failed", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleBrowseNamespace(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool, namespace string) {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "namespace is required", nil)
		return
	}

	// Ensure namespace exists.
	if !namespaceExists(req.Context(), db, namespace) {
		writeError(w, http.StatusNotFound, "not_found", "namespace not found", nil)
		return
	}

	prefix, err := parsePrefix(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}
	limit, err := parseLimit(req, 200)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}
	offset, err := parseCursorOffset(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}

	startIndex := len(prefix) + 1 // SQL substr is 1-based

	rows, err := db.Query(req.Context(), `
		WITH matches AS (
			SELECT
				c.path,
				c.format::text AS format,
				lv.version AS latest_version
			FROM configs c
			LEFT JOIN LATERAL (
				SELECT version
				FROM config_versions
				WHERE config_id = c.id
				ORDER BY version DESC
				LIMIT 1
			) lv ON true
			WHERE c.namespace = $1
			  AND ($2 = '' OR c.path LIKE $2 || '%')
		),
		agg AS (
			SELECT
				split_part(substr(path, $3), '/', 1) AS child,
				bool_or(position('/' in substr(path, $3)) > 0) AS has_folder,
				bool_or(position('/' in substr(path, $3)) = 0) AS has_config,
				max(CASE WHEN position('/' in substr(path, $3)) = 0 THEN format END) AS leaf_format,
				max(CASE WHEN position('/' in substr(path, $3)) = 0 THEN latest_version END) AS leaf_latest_version
			FROM matches
			GROUP BY child
		)
		SELECT child, has_folder, has_config, leaf_format, leaf_latest_version
		FROM agg
		ORDER BY child ASC
		LIMIT $4 OFFSET $5
	`, namespace, prefix, startIndex, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}
	defer rows.Close()

	entries := make([]any, 0, limit)
	for rows.Next() {
		var child string
		var hasFolder, hasConfig bool
		var leafFormat sql.NullString
		var leafLatest sql.NullInt32
		if err := rows.Scan(&child, &hasFolder, &hasConfig, &leafFormat, &leafLatest); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "scan failed", nil)
			return
		}
		if child == "" {
			continue
		}

		if hasFolder {
			full := prefix + child + "/"
			entries = append(entries, BrowseEntryFolder{
				Type:     "folder",
				Name:     child,
				FullPath: full,
			})
		}
		if hasConfig {
			if !leafFormat.Valid || !leafLatest.Valid {
				// Defensive: a config row should always have format and a latest version.
				continue
			}
			full := prefix + child
			entries = append(entries, BrowseEntryConfig{
				Type:          "config",
				Name:          child,
				FullPath:      full,
				Format:        ConfigFormat(leafFormat.String),
				LatestVersion: int(leafLatest.Int32),
			})
		}
	}

	var next *string
	if len(entries) >= limit {
		c := encodeCursorOffset(offset + limit)
		next = &c
	}

	writeJSON(w, http.StatusOK, BrowseResponse{
		Items:      entries,
		NextCursor: next,
	})
}

func namespaceExists(ctx context.Context, db *pgxpool.Pool, name string) bool {
	var ok bool
	_ = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM namespaces WHERE name = $1)`, name).Scan(&ok)
	return ok
}

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return u.String()
}
