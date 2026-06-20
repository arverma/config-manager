-- Authentication: sessions, OAuth state, and API keys

CREATE TABLE IF NOT EXISTS auth_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  token_hash TEXT NOT NULL,
  actor_type TEXT NOT NULL,
  actor_id TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_activity_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,

  CONSTRAINT auth_sessions_token_hash_unique UNIQUE (token_hash),
  CONSTRAINT auth_sessions_actor_type_nonempty CHECK (char_length(trim(actor_type)) > 0),
  CONSTRAINT auth_sessions_actor_id_nonempty CHECK (char_length(trim(actor_id)) > 0)
);

CREATE INDEX IF NOT EXISTS auth_sessions_expires_at_idx ON auth_sessions (expires_at);

CREATE TABLE IF NOT EXISTS auth_oauth_states (
  state TEXT PRIMARY KEY,
  return_to TEXT NOT NULL DEFAULT '/',
  expires_at TIMESTAMPTZ NOT NULL,

  CONSTRAINT auth_oauth_states_state_nonempty CHECK (char_length(trim(state)) > 0)
);

CREATE INDEX IF NOT EXISTS auth_oauth_states_expires_at_idx ON auth_oauth_states (expires_at);

CREATE TABLE IF NOT EXISTS api_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  key_hash TEXT NOT NULL,
  key_prefix TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_used_at TIMESTAMPTZ NULL,
  revoked_at TIMESTAMPTZ NULL,

  CONSTRAINT api_keys_name_unique UNIQUE (name),
  CONSTRAINT api_keys_name_nonempty CHECK (char_length(trim(name)) > 0),
  CONSTRAINT api_keys_key_hash_unique UNIQUE (key_hash),
  CONSTRAINT api_keys_key_prefix_nonempty CHECK (char_length(trim(key_prefix)) > 0)
);

CREATE INDEX IF NOT EXISTS api_keys_revoked_at_idx ON api_keys (revoked_at);
