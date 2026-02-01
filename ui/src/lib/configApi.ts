export type ConfigFormat = "json" | "yaml";

export function getConfigApiBaseUrl(): string {
  // Server-side can read non-public env vars too.
  return (
    process.env.CONFIG_API_BASE_URL ||
    process.env.NEXT_PUBLIC_CONFIG_API_BASE_URL ||
    "http://localhost:8080"
  );
}

export function buildConfigUrl(args: {
  baseUrl: string;
  namespace: string;
  path: string;
}): string {
  const ns = encodeURIComponent(args.namespace);
  // `path` is folder-like; keep slashes to mirror REST URL.
  return `${args.baseUrl}/configs/${ns}/${args.path}`;
}

export function buildNamespacesUrl(args: { baseUrl: string }): string {
  return `${args.baseUrl}/namespaces`;
}

export function buildNamespaceBrowseUrl(args: {
  baseUrl: string;
  namespace: string;
  prefix?: string;
}): string {
  const qp = new URLSearchParams();
  if (args.prefix) qp.set("prefix", args.prefix);
  return `${args.baseUrl}/namespaces/${encodeURIComponent(args.namespace)}/browse?${qp.toString()}`;
}

export function buildVersionsUrl(args: {
  baseUrl: string;
  namespace: string;
  path: string;
}): string {
  return `${args.baseUrl}/configs/${encodeURIComponent(args.namespace)}/${args.path}/versions`;
}

export function buildVersionUrl(args: {
  baseUrl: string;
  namespace: string;
  path: string;
  version: number;
}): string {
  return `${args.baseUrl}/configs/${encodeURIComponent(args.namespace)}/${args.path}/versions/${args.version}`;
}

export function buildDeleteVersionUrl(args: {
  baseUrl: string;
  namespace: string;
  path: string;
  version: number;
}): string {
  // Same path as buildVersionUrl, but used with DELETE.
  return buildVersionUrl(args);
}

export function buildDeleteConfigUrl(args: {
  baseUrl: string;
  namespace: string;
  path: string;
}): string {
  return `${args.baseUrl}/configs/${encodeURIComponent(args.namespace)}/${args.path}`;
}

export function buildDeleteNamespaceUrl(args: {
  baseUrl: string;
  namespace: string;
}): string {
  return `${args.baseUrl}/namespaces/${encodeURIComponent(args.namespace)}`;
}

