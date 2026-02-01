package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func storeGetConfigOnly(ctx context.Context, q querier, namespace, path string) (Config, pgtype.UUID, error) {
	var cfgID pgtype.UUID
	var cfg Config
	var fmtStr string
	err := q.QueryRow(ctx, `
		SELECT id, namespace, path, format::text, created_at, updated_at
		FROM configs
		WHERE namespace = $1 AND path = $2
		  AND deleted_at IS NULL
	`, namespace, path).Scan(&cfgID, &cfg.Namespace, &cfg.Path, &fmtStr, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		return Config{}, pgtype.UUID{}, err
	}
	cfg.ID = uuidToString(cfgID)
	cfg.Format = ConfigFormat(fmtStr)
	return cfg, cfgID, nil
}

func storeGetConfigAndLatest(ctx context.Context, q querier, namespace, path string) (Config, ConfigVersion, error) {
	cfg, cfgID, err := storeGetConfigOnly(ctx, q, namespace, path)
	if err != nil {
		return Config{}, ConfigVersion{}, err
	}
	ver, err := storeGetLatestVersion(ctx, q, cfgID)
	if err != nil {
		return Config{}, ConfigVersion{}, err
	}
	cfg.LatestVersionID = ptr(ver.ID)
	return cfg, ver, nil
}

func storeGetLatestVersion(ctx context.Context, q querier, cfgID pgtype.UUID) (ConfigVersion, error) {
	row := q.QueryRow(ctx, `
		SELECT id, version, created_at, created_by, comment, content_sha256, body_raw, body_json
		FROM config_versions
		WHERE config_id = $1
		ORDER BY version DESC
		LIMIT 1
	`, cfgID)
	return scanConfigVersion(row)
}

func storeGetVersion(ctx context.Context, q querier, cfgID pgtype.UUID, version int) (ConfigVersion, error) {
	row := q.QueryRow(ctx, `
		SELECT id, version, created_at, created_by, comment, content_sha256, body_raw, body_json
		FROM config_versions
		WHERE config_id = $1 AND version = $2
	`, cfgID, version)
	return scanConfigVersion(row)
}

func scanConfigVersion(s rowScanner) (ConfigVersion, error) {
	var verID pgtype.UUID
	var v ConfigVersion
	var bodyJSON []byte
	var createdBy, comment, contentSHA sql.NullString
	if err := s.Scan(&verID, &v.Version, &v.CreatedAt, &createdBy, &comment, &contentSHA, &v.BodyRaw, &bodyJSON); err != nil {
		return ConfigVersion{}, err
	}

	v.ID = uuidToString(verID)
	if createdBy.Valid {
		v.CreatedBy = &createdBy.String
	}
	if comment.Valid {
		v.Comment = &comment.String
	}
	if contentSHA.Valid {
		v.ContentSHA256 = &contentSHA.String
	}
	if bodyJSON != nil {
		var anyVal any
		_ = json.Unmarshal(bodyJSON, &anyVal)
		v.BodyJSON = anyVal
	}
	return v, nil
}

func scanConfigVersionMeta(s rowScanner) (ConfigVersionMeta, error) {
	var id pgtype.UUID
	var m ConfigVersionMeta
	var createdBy, comment, contentSHA sql.NullString
	if err := s.Scan(&id, &m.Version, &m.CreatedAt, &createdBy, &comment, &contentSHA); err != nil {
		return ConfigVersionMeta{}, err
	}
	m.ID = uuidToString(id)
	if createdBy.Valid {
		m.CreatedBy = &createdBy.String
	}
	if comment.Valid {
		m.Comment = &comment.String
	}
	if contentSHA.Valid {
		m.ContentSHA256 = &contentSHA.String
	}
	return m, nil
}
