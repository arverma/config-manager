export type ConfigFormat = "json" | "yaml";

/** Base URL for API calls: /api in browser (proxy Route Handler or ingress); server-side uses env. */
export function getConfigApiBaseUrl(): string {
  if (typeof window !== "undefined") {
    return "/api";
  }
  return (
    process.env.CONFIG_API_BASE_URL ||
    process.env.NEXT_PUBLIC_CONFIG_API_BASE_URL ||
    "http://localhost:8080"
  );
}

function encodePathPreservingSlashes(path: string): string {
  return path
    .split("/")
    .map((seg) => encodeURIComponent(seg))
    .join("/");
}

export function buildConfigPath(args: { namespace: string; path: string }): string {
  const ns = encodeURIComponent(args.namespace);
  const p = encodePathPreservingSlashes(args.path);
  return `/configs/${ns}/${p}`;
}

export function buildNamespacesPath(): string {
  return "/namespaces";
}

export function buildNamespaceBrowsePath(args: {
  namespace: string;
  prefix?: string;
}): string {
  const qp = new URLSearchParams();
  if (args.prefix) qp.set("prefix", args.prefix);
  const q = qp.toString();
  return `/namespaces/${encodeURIComponent(args.namespace)}/browse${q ? `?${q}` : ""}`;
}

export function buildVersionsPath(args: { namespace: string; path: string }): string {
  const ns = encodeURIComponent(args.namespace);
  const p = encodePathPreservingSlashes(args.path);
  return `/configs/${ns}/${p}/versions`;
}

export function buildVersionPath(args: {
  namespace: string;
  path: string;
  version: number;
}): string {
  const ns = encodeURIComponent(args.namespace);
  const p = encodePathPreservingSlashes(args.path);
  return `/configs/${ns}/${p}/versions/${args.version}`;
}

export function buildDeleteVersionPath(args: {
  namespace: string;
  path: string;
  version: number;
}): string {
  return buildVersionPath(args);
}

export function buildDeleteConfigPath(args: {
  namespace: string;
  path: string;
}): string {
  return buildConfigPath(args);
}

export function buildDeleteNamespacePath(args: { namespace: string }): string {
  return `/namespaces/${encodeURIComponent(args.namespace)}`;
}

