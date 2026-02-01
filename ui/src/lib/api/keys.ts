export const queryKeys = {
  namespaces: () => ["namespaces"] as const,
  namespaceBrowse: (namespace: string, prefix: string) =>
    ["namespaceBrowse", namespace, prefix] as const,
  configLatest: (namespace: string, path: string) =>
    ["configLatest", namespace, path] as const,
  configVersions: (namespace: string, path: string) =>
    ["configVersions", namespace, path] as const,
  configVersion: (namespace: string, path: string, version: number) =>
    ["configVersion", namespace, path, version] as const,
};

