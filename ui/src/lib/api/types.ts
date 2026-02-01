import type { ConfigFormat } from "@/lib/configApi";

export type Namespace = {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
};

export type NamespaceWithCount = Namespace & {
  config_count: number;
};

export type NamespaceListResponse = {
  items: NamespaceWithCount[];
  next_cursor?: string | null;
};

export type Config = {
  id: string;
  namespace: string;
  path: string;
  format: ConfigFormat;
  latest_version_id?: string;
  created_at: string;
  updated_at: string;
};

export type ConfigVersion = {
  id: string;
  version: number;
  created_at: string;
  created_by?: string;
  comment?: string;
  content_sha256?: string;
  body_raw: string;
  body_json?: unknown;
};

export type ConfigVersionMeta = Omit<ConfigVersion, "body_raw" | "body_json">;

export type GetConfigResponse = {
  config: Config;
  latest: ConfigVersion;
};

export type GetVersionResponse = {
  config: Config;
  version: ConfigVersion;
};

export type VersionListResponse = {
  items: ConfigVersionMeta[];
  next_cursor?: string | null;
};

export type BrowseEntry =
  | { type: "folder"; name: string; full_path: string }
  | {
      type: "config";
      name: string;
      full_path: string;
      format: ConfigFormat;
      latest_version: number;
    };

export type BrowseResponse = {
  items: BrowseEntry[];
  next_cursor?: string | null;
};

