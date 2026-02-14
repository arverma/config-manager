-- Restore original namespace name pattern (lowercase and underscore only).
ALTER TABLE namespaces DROP CONSTRAINT IF EXISTS namespaces_name_pattern;
ALTER TABLE namespaces ADD CONSTRAINT namespaces_name_pattern CHECK (name ~ '^[a-z_]+$');
