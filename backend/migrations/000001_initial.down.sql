-- Rollback initial schema: drop triggers, tables, type, functions.
-- Does not drop pgcrypto extension (may be used elsewhere).

DROP TRIGGER IF EXISTS configs_enforce_latest_version ON configs;
DROP TRIGGER IF EXISTS configs_set_updated_at ON configs;
DROP TRIGGER IF EXISTS namespaces_set_updated_at ON namespaces;

DROP TABLE IF EXISTS config_versions;
DROP TABLE IF EXISTS configs;
DROP TABLE IF EXISTS namespaces;

DROP TYPE IF EXISTS config_format;

DROP FUNCTION IF EXISTS enforce_latest_version_belongs_to_config();
DROP FUNCTION IF EXISTS set_updated_at();
