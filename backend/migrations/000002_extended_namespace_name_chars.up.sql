-- Allow namespace names: letters (a-z, A-Z), digits, underscore, hyphen.
ALTER TABLE namespaces DROP CONSTRAINT IF EXISTS namespaces_name_pattern;
ALTER TABLE namespaces ADD CONSTRAINT namespaces_name_pattern CHECK (name ~ '^[a-zA-Z0-9_-]+$');
