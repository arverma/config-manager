package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxConfigBodyBytes = int64(5 << 20) // 5 MiB
)

func handleListConfigs(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool) {
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
	namespace := strings.TrimSpace(req.URL.Query().Get("namespace"))
	prefix, err := parsePrefix(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}
	recursive, hasRecursive, err := parseOptionalBool(req, "recursive")
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}
	if !hasRecursive {
		recursive = true
	}

	// Basic list. Vault-like folder browsing is handled by /namespaces/{namespace}/browse.
	var rows pgx.Rows
	if recursive {
		rows, err = db.Query(req.Context(), `
			SELECT
				c.id, c.namespace, c.path, c.format::text, c.created_at, c.updated_at,
				lv.id, lv.version, lv.created_at, lv.created_by, lv.comment, lv.content_sha256
			FROM configs c
			LEFT JOIN LATERAL (
				SELECT id, version, created_at, created_by, comment, content_sha256
				FROM config_versions
				WHERE config_id = c.id
				ORDER BY version DESC
				LIMIT 1
			) lv ON true
			WHERE ($1 = '' OR c.namespace = $1)
			  AND ($2 = '' OR c.path LIKE $2 || '%')
			  AND c.deleted_at IS NULL
			ORDER BY c.namespace ASC, c.path ASC
			LIMIT $3 OFFSET $4
		`, namespace, prefix, limit, offset)
	} else {
		startIndex := len(prefix) + 1 // SQL substr is 1-based
		rows, err = db.Query(req.Context(), `
			SELECT
				c.id, c.namespace, c.path, c.format::text, c.created_at, c.updated_at,
				lv.id, lv.version, lv.created_at, lv.created_by, lv.comment, lv.content_sha256
			FROM configs c
			LEFT JOIN LATERAL (
				SELECT id, version, created_at, created_by, comment, content_sha256
				FROM config_versions
				WHERE config_id = c.id
				ORDER BY version DESC
				LIMIT 1
			) lv ON true
			WHERE ($1 = '' OR c.namespace = $1)
			  AND ($2 = '' OR c.path LIKE $2 || '%')
			  AND c.deleted_at IS NULL
			  AND position('/' in substr(c.path, $3)) = 0
			ORDER BY c.namespace ASC, c.path ASC
			LIMIT $4 OFFSET $5
		`, namespace, prefix, startIndex, limit, offset)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}
	defer rows.Close()

	items := make([]ConfigListItem, 0, limit)
	for rows.Next() {
		var cfgID, latestVerID pgtype.UUID
		var cfg Config
		var fmtStr string
		var latestMeta ConfigVersionMeta
		var latestCreatedAt pgtype.Timestamptz
		var createdBy, comment, contentSHA sql.NullString
		if err := rows.Scan(
			&cfgID, &cfg.Namespace, &cfg.Path, &fmtStr, &cfg.CreatedAt, &cfg.UpdatedAt,
			&latestVerID, &latestMeta.Version, &latestCreatedAt, &createdBy, &comment, &contentSHA,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "scan failed", nil)
			return
		}
		cfg.ID = uuidToString(cfgID)
		cfg.Format = ConfigFormat(fmtStr)
		latestMeta.CreatedAt = latestCreatedAt.Time
		latestMeta.ID = uuidToString(latestVerID)
		cfg.LatestVersionID = &latestMeta.ID
		if createdBy.Valid {
			latestMeta.CreatedBy = &createdBy.String
		}
		if comment.Valid {
			latestMeta.Comment = &comment.String
		}
		if contentSHA.Valid {
			latestMeta.ContentSHA256 = &contentSHA.String
		}
		items = append(items, ConfigListItem{Config: cfg, LatestMeta: latestMeta})
	}

	var next *string
	if len(items) == limit {
		c := encodeCursorOffset(offset + limit)
		next = &c
	}

	writeJSON(w, http.StatusOK, ConfigListResponse{Items: items, NextCursor: next})
}

func handleGetLatestConfig(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool) {
	namespace, path, ok := getNamespaceAndPath(w, req)
	if !ok {
		return
	}

	cfg, ver, err := storeGetConfigAndLatest(req.Context(), db, namespace, path)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "config not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}

	writeJSON(w, http.StatusOK, GetConfigResponse{Config: cfg, Latest: ver})
}

func handleCreateConfig(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool) {
	namespace, path, ok := getNamespaceAndPath(w, req)
	if !ok {
		return
	}

	var body struct {
		Format    ConfigFormat `json:"format"`
		BodyRaw   string       `json:"body_raw"`
		Comment   *string      `json:"comment"`
		CreatedBy *string      `json:"created_by"`
	}
	if err := decodeJSONBody(w, req, &body, maxConfigBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}
	if body.BodyRaw == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "body_raw is required", map[string]any{"field": "body_raw"})
		return
	}
	if body.Format != FormatJSON && body.Format != FormatYAML {
		writeError(w, http.StatusBadRequest, "bad_request", "format must be one of: json, yaml", map[string]any{"field": "format"})
		return
	}

	parsedAny, parsedJSON, err := parseBody(body.Format, body.BodyRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}
	sha := sha256Hex(body.BodyRaw)
	reqID, userAgent, sourceIP := requestAuditFields(req)

	nsOK, err := storeNamespaceExists(req.Context(), db, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}
	if !nsOK {
		writeError(w, http.StatusNotFound, "not_found", "namespace not found", nil)
		return
	}

	tx, err := db.BeginTx(req.Context(), pgx.TxOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "begin failed", nil)
		return
	}
	defer tx.Rollback(req.Context())

	var cfgID pgtype.UUID
	var latestVersionID pgtype.UUID
	var createdAt, updatedAt pgtype.Timestamptz

	// Insert config (namespace must exist; FK enforces).
	err = tx.QueryRow(req.Context(), `
		INSERT INTO configs (namespace, path, format)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, namespace, path, string(body.Format)).Scan(&cfgID, &createdAt, &updatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				writeError(w, http.StatusConflict, "conflict", "config already exists", nil)
				return
			case "23503":
				writeError(w, http.StatusNotFound, "not_found", "namespace not found", nil)
				return
			}
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "insert failed", nil)
		return
	}

	// Insert version 1
	var versionCreatedAt pgtype.Timestamptz
	err = tx.QueryRow(req.Context(), `
		INSERT INTO config_versions (config_id, version, body_raw, body_json, created_by, comment, content_sha256, request_id, user_agent, source_ip)
		VALUES ($1, 1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`, cfgID, body.BodyRaw, json.RawMessage(parsedJSON), body.CreatedBy, body.Comment, sha, reqID, userAgent, sourceIP).Scan(&latestVersionID, &versionCreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "insert version failed", nil)
		return
	}

	// Update latest pointer
	_, err = tx.Exec(req.Context(), `UPDATE configs SET latest_version_id = $1 WHERE id = $2`, latestVersionID, cfgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "update latest failed", nil)
		return
	}

	if err := tx.Commit(req.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "commit failed", nil)
		return
	}

	cfg := Config{
		ID:              uuidToString(cfgID),
		Namespace:       namespace,
		Path:            path,
		Format:          body.Format,
		LatestVersionID: ptr(uuidToString(latestVersionID)),
		CreatedAt:       createdAt.Time,
		UpdatedAt:       updatedAt.Time,
	}
	ver := ConfigVersion{
		ID:            uuidToString(latestVersionID),
		Version:       1,
		CreatedAt:     versionCreatedAt.Time,
		CreatedBy:     body.CreatedBy,
		Comment:       body.Comment,
		ContentSHA256: ptr(sha),
		BodyRaw:       body.BodyRaw,
		BodyJSON:      parsedAny,
	}
	writeJSON(w, http.StatusCreated, GetConfigResponse{Config: cfg, Latest: ver})
}

func handleUpdateConfig(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool) {
	namespace, path, ok := getNamespaceAndPath(w, req)
	if !ok {
		return
	}

	var body struct {
		BodyRaw     string  `json:"body_raw"`
		Comment     *string `json:"comment"`
		CreatedBy   *string `json:"created_by"`
		BaseVersion *int    `json:"base_version"`
	}
	if err := decodeJSONBody(w, req, &body, maxConfigBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}
	if body.BodyRaw == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "body_raw is required", map[string]any{"field": "body_raw"})
		return
	}

	tx, err := db.BeginTx(req.Context(), pgx.TxOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "begin failed", nil)
		return
	}
	defer tx.Rollback(req.Context())

	// Lock config row to ensure version increments safely.
	var cfgID, latestVersionID pgtype.UUID
	var cfg Config
	var fmtStr string
	err = tx.QueryRow(req.Context(), `
		SELECT c.id, c.namespace, c.path, c.format::text, c.latest_version_id, c.created_at, c.updated_at
		FROM configs c
		WHERE c.namespace = $1 AND c.path = $2
		  AND c.deleted_at IS NULL
		FOR UPDATE
	`, namespace, path).Scan(&cfgID, &cfg.Namespace, &cfg.Path, &fmtStr, &latestVersionID, &cfg.CreatedAt, &cfg.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "config not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}
	cfg.ID = uuidToString(cfgID)
	cfg.Format = ConfigFormat(fmtStr)
	if latestVersionID.Valid {
		s := uuidToString(latestVersionID)
		cfg.LatestVersionID = &s
	}

	// Latest is strictly the max(version).
	var currentLatestNumber int
	if err := tx.QueryRow(req.Context(), `SELECT COALESCE(MAX(version), 0) FROM config_versions WHERE config_id = $1`, cfgID).Scan(&currentLatestNumber); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}

	if body.BaseVersion != nil && *body.BaseVersion != currentLatestNumber {
		writeError(w, http.StatusConflict, "conflict", "base_version does not match current latest", map[string]any{
			"base_version":    *body.BaseVersion,
			"current_version": currentLatestNumber,
		})
		return
	}

	// No-op guard: if submitted body matches current latest exactly, do not create a new version.
	// This keeps version history meaningful and prevents accidental duplicate versions.
	sha := sha256Hex(body.BodyRaw)
	if currentLatestNumber > 0 {
		var latestSHA sql.NullString
		var latestBodyRaw string
		err := tx.QueryRow(req.Context(), `
			SELECT content_sha256, body_raw
			FROM config_versions
			WHERE config_id = $1 AND version = $2
		`, cfgID, currentLatestNumber).Scan(&latestSHA, &latestBodyRaw)
		if err == nil {
			latest := ""
			if latestSHA.Valid {
				latest = latestSHA.String
			}
			if latest == "" {
				latest = sha256Hex(latestBodyRaw)
			}
			if latest == sha {
				writeError(w, http.StatusConflict, "no_change", "body_raw matches current latest", map[string]any{
					"current_version": currentLatestNumber,
				})
				return
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
			return
		}
	}

	parsedAny, parsedJSON, err := parseBody(cfg.Format, body.BodyRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}
	reqID, userAgent, sourceIP := requestAuditFields(req)

	nextVersion := currentLatestNumber + 1
	var newVerID pgtype.UUID
	var createdAt pgtype.Timestamptz
	err = tx.QueryRow(req.Context(), `
		INSERT INTO config_versions (config_id, version, body_raw, body_json, created_by, comment, content_sha256, request_id, user_agent, source_ip)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`, cfgID, nextVersion, body.BodyRaw, json.RawMessage(parsedJSON), body.CreatedBy, body.Comment, sha, reqID, userAgent, sourceIP).Scan(&newVerID, &createdAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "insert version failed", nil)
		return
	}
	_, err = tx.Exec(req.Context(), `UPDATE configs SET latest_version_id = $1 WHERE id = $2`, newVerID, cfgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "update latest failed", nil)
		return
	}

	if err := tx.Commit(req.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "commit failed", nil)
		return
	}

	cfg.LatestVersionID = ptr(uuidToString(newVerID))
	ver := ConfigVersion{
		ID:            uuidToString(newVerID),
		Version:       nextVersion,
		CreatedAt:     createdAt.Time,
		CreatedBy:     body.CreatedBy,
		Comment:       body.Comment,
		ContentSHA256: ptr(sha),
		BodyRaw:       body.BodyRaw,
		BodyJSON:      parsedAny,
	}
	writeJSON(w, http.StatusOK, GetConfigResponse{Config: cfg, Latest: ver})
}

func handleListConfigVersions(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool) {
	namespace, path, ok := getNamespaceAndPath(w, req)
	if !ok {
		return
	}
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

	_, cfgID, err := storeGetConfigOnly(req.Context(), db, namespace, path)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "config not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}

	rows, err := db.Query(req.Context(), `
		SELECT id, version, created_at, created_by, comment, content_sha256
		FROM config_versions
		WHERE config_id = $1
		ORDER BY version DESC
		LIMIT $2 OFFSET $3
	`, cfgID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}
	defer rows.Close()

	items := make([]ConfigVersionMeta, 0, limit)
	for rows.Next() {
		m, err := scanConfigVersionMeta(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "scan failed", nil)
			return
		}
		items = append(items, m)
	}

	var next *string
	if len(items) == limit {
		c := encodeCursorOffset(offset + limit)
		next = &c
	}
	writeJSON(w, http.StatusOK, VersionListResponse{Items: items, NextCursor: next})
}

func handleGetConfigVersion(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool) {
	namespace, path, ok := getNamespaceAndPath(w, req)
	if !ok {
		return
	}
	verNumStr := chi.URLParam(req, "version")
	verNum, _ := strconv.Atoi(verNumStr)

	cfg, cfgID, err := storeGetConfigOnly(req.Context(), db, namespace, path)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "config not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}

	ver, err := storeGetVersion(req.Context(), db, cfgID, verNum)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "version not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}

	// Derive latest pointer as max(version) for this config.
	latest, err := storeGetLatestVersion(req.Context(), db, cfgID)
	if err == nil {
		cfg.LatestVersionID = ptr(latest.ID)
	}

	writeJSON(w, http.StatusOK, GetVersionResponse{Config: cfg, Version: ver})
}

func handleDeleteConfigVersion(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool) {
	namespace, path, ok := getNamespaceAndPath(w, req)
	if !ok {
		return
	}
	verNumStr := chi.URLParam(req, "version")
	verNum, _ := strconv.Atoi(verNumStr)

	tx, err := db.BeginTx(req.Context(), pgx.TxOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "begin failed", nil)
		return
	}
	defer tx.Rollback(req.Context())

	// Lock config row.
	var cfgID pgtype.UUID
	err = tx.QueryRow(req.Context(), `
		SELECT id
		FROM configs
		WHERE namespace = $1 AND path = $2
		  AND deleted_at IS NULL
		FOR UPDATE
	`, namespace, path).Scan(&cfgID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "config not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}

	// Determine latest version number.
	var latestNum int
	if err := tx.QueryRow(req.Context(), `SELECT COALESCE(MAX(version), 0) FROM config_versions WHERE config_id = $1`, cfgID).Scan(&latestNum); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}
	if verNum == latestNum {
		writeError(w, http.StatusConflict, "conflict", "cannot delete latest version", map[string]any{"latest_version": latestNum})
		return
	}

	tag, err := tx.Exec(req.Context(), `
		DELETE FROM config_versions
		WHERE config_id = $1 AND version = $2
	`, cfgID, verNum)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "delete failed", nil)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "not_found", "version not found", nil)
		return
	}

	if err := tx.Commit(req.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "commit failed", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteConfig(w http.ResponseWriter, req *http.Request, db *pgxpool.Pool) {
	namespace, path, ok := getNamespaceAndPath(w, req)
	if !ok {
		return
	}

	tx, err := db.BeginTx(req.Context(), pgx.TxOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "begin failed", nil)
		return
	}
	defer tx.Rollback(req.Context())

	// Lock config row.
	var cfgID pgtype.UUID
	err = tx.QueryRow(req.Context(), `
		SELECT id
		FROM configs
		WHERE namespace = $1 AND path = $2
		  AND deleted_at IS NULL
		FOR UPDATE
	`, namespace, path).Scan(&cfgID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "config not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "query failed", nil)
		return
	}

	tag, err := tx.Exec(req.Context(), `DELETE FROM configs WHERE id = $1`, cfgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "delete failed", nil)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "not_found", "config not found", nil)
		return
	}

	if err := tx.Commit(req.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "commit failed", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getNamespaceAndPath(w http.ResponseWriter, req *http.Request) (string, string, bool) {
	namespace := strings.TrimSpace(chi.URLParam(req, "namespace"))
	if err := validateNamespace(namespace); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return "", "", false
	}
	namespace = strings.TrimSpace(namespace)

	path, err := normalizeConfigPath(chi.URLParam(req, "path"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return "", "", false
	}
	return namespace, path, true
}
