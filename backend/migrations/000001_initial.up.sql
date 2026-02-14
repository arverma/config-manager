-- Config Manager - initial schema (versioned migration)
-- Identity model: (namespace, path) identifies a config; config_versions are immutable.
-- Postgres 13+. Source of truth: backend/migrations/; API runs these on startup.

-- UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Enumerated format ("type" in your requirements)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'config_format') THEN
    CREATE TYPE config_format AS ENUM ('json', 'yaml');
  END IF;
END$$;

-- Updated-at helper
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Explicit namespaces (globally unique)
CREATE TABLE IF NOT EXISTS namespaces (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  name TEXT NOT NULL,

  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT namespaces_name_nonempty CHECK (char_length(trim(name)) > 0),
  CONSTRAINT namespaces_name_pattern CHECK (name ~ '^[a-z_]+$'),

  CONSTRAINT namespaces_name_unique UNIQUE (name)
);

CREATE TRIGGER namespaces_set_updated_at
BEFORE UPDATE ON namespaces
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Logical config
CREATE TABLE IF NOT EXISTS configs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  namespace TEXT NOT NULL,
  path      TEXT NOT NULL,
  format    config_format NOT NULL,

  latest_version_id UUID NULL,

  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  deleted_at TIMESTAMPTZ NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT configs_namespace_nonempty CHECK (char_length(trim(namespace)) > 0),
  CONSTRAINT configs_path_nonempty CHECK (char_length(trim(path)) > 0),
  CONSTRAINT configs_namespace_no_slash CHECK (position('/' in namespace) = 0),
  CONSTRAINT configs_path_no_leading_slash CHECK (path !~ '^/'),
  CONSTRAINT configs_path_no_trailing_slash CHECK (path !~ '/$'),
  CONSTRAINT configs_path_no_empty_segments CHECK (path !~ '//'),
  CONSTRAINT configs_path_no_parent_segments CHECK (path !~ '(^|/)\\.\\.(\\/|$)'),
  CONSTRAINT configs_path_no_whitespace CHECK (path !~ '[[:space:]]'),

  CONSTRAINT configs_namespace_fk
    FOREIGN KEY (namespace) REFERENCES namespaces(name)
);

CREATE TRIGGER configs_set_updated_at
BEFORE UPDATE ON configs
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Immutable versions
CREATE TABLE IF NOT EXISTS config_versions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  config_id UUID NOT NULL REFERENCES configs(id) ON DELETE CASCADE,
  version   INTEGER NOT NULL,

  body_raw  TEXT NOT NULL,
  body_json JSONB NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by TEXT NULL,
  comment    TEXT NULL,
  content_sha256 TEXT NULL,
  request_id TEXT NULL,
  user_agent TEXT NULL,
  source_ip  INET NULL,

  CONSTRAINT config_versions_version_positive CHECK (version >= 1),
  CONSTRAINT config_versions_unique_per_config UNIQUE (config_id, version)
);

CREATE INDEX IF NOT EXISTS config_versions_config_id_idx
  ON config_versions (config_id);

-- Add FK for configs.latest_version_id after config_versions exists
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.table_constraints tc
    WHERE tc.table_name = 'configs'
      AND tc.constraint_type = 'FOREIGN KEY'
      AND tc.constraint_name = 'configs_latest_version_fk'
  ) THEN
    ALTER TABLE configs
      ADD CONSTRAINT configs_latest_version_fk
      FOREIGN KEY (latest_version_id) REFERENCES config_versions(id);
  END IF;
END$$;

-- Ensure latest_version_id (when set) points to a version for the same config.
CREATE OR REPLACE FUNCTION enforce_latest_version_belongs_to_config()
RETURNS TRIGGER AS $$
DECLARE
  ok BOOLEAN;
BEGIN
  IF NEW.latest_version_id IS NULL THEN
    RETURN NEW;
  END IF;

  SELECT EXISTS (
    SELECT 1
    FROM config_versions cv
    WHERE cv.id = NEW.latest_version_id
      AND cv.config_id = NEW.id
  ) INTO ok;

  IF NOT ok THEN
    RAISE EXCEPTION 'latest_version_id must reference a version belonging to the same config';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS configs_enforce_latest_version ON configs;
CREATE TRIGGER configs_enforce_latest_version
BEFORE INSERT OR UPDATE OF latest_version_id ON configs
FOR EACH ROW
EXECUTE FUNCTION enforce_latest_version_belongs_to_config();

-- Browse/index helpers (UI listing and prefix filtering)
CREATE UNIQUE INDEX IF NOT EXISTS configs_identity_active_unique
  ON configs (namespace, path)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS configs_browse_active_idx
  ON configs (namespace, path)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS configs_path_prefix_active_idx
  ON configs (namespace, path text_pattern_ops)
  WHERE deleted_at IS NULL;
