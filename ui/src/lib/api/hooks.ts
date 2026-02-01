import { useQuery, type QueryClient } from "@tanstack/react-query";

import { apiFetch } from "@/lib/api/client";
import { queryKeys } from "@/lib/api/keys";
import type {
  BrowseResponse,
  GetConfigResponse,
  GetVersionResponse,
  NamespaceListResponse,
  VersionListResponse,
} from "@/lib/api/types";
import {
  buildConfigUrl,
  buildNamespaceBrowseUrl,
  buildNamespacesUrl,
  buildVersionUrl,
  buildVersionsUrl,
} from "@/lib/configApi";

export function namespacesQueryOptions(args: { baseUrl: string; limit?: number }) {
  const url = buildNamespacesUrl({ baseUrl: args.baseUrl });
  const qp = new URLSearchParams();
  if (args.limit) qp.set("limit", String(args.limit));
  const fullUrl = qp.toString() ? `${url}?${qp.toString()}` : url;

  return {
    queryKey: queryKeys.namespaces(),
    queryFn: async (): Promise<NamespaceListResponse> =>
      await apiFetch<NamespaceListResponse>(fullUrl),
  };
}

export function namespaceBrowseQueryOptions(args: {
  baseUrl: string;
  namespace: string;
  prefix: string;
}) {
  const url = buildNamespaceBrowseUrl({
    baseUrl: args.baseUrl,
    namespace: args.namespace,
    prefix: args.prefix,
  });
  return {
    queryKey: queryKeys.namespaceBrowse(args.namespace, args.prefix),
    queryFn: async (): Promise<BrowseResponse> => await apiFetch<BrowseResponse>(url),
  };
}

export function configLatestQueryOptions(args: {
  baseUrl: string;
  namespace: string;
  path: string;
}) {
  const url = buildConfigUrl({
    baseUrl: args.baseUrl,
    namespace: args.namespace,
    path: args.path,
  });
  return {
    queryKey: queryKeys.configLatest(args.namespace, args.path),
    queryFn: async (): Promise<GetConfigResponse> =>
      await apiFetch<GetConfigResponse>(url),
  };
}

export function configVersionsQueryOptions(args: {
  baseUrl: string;
  namespace: string;
  path: string;
}) {
  const url = buildVersionsUrl({
    baseUrl: args.baseUrl,
    namespace: args.namespace,
    path: args.path,
  });
  return {
    queryKey: queryKeys.configVersions(args.namespace, args.path),
    queryFn: async (): Promise<VersionListResponse> =>
      await apiFetch<VersionListResponse>(url),
  };
}

export function configVersionQueryOptions(args: {
  baseUrl: string;
  namespace: string;
  path: string;
  version: number;
}) {
  const url = buildVersionUrl({
    baseUrl: args.baseUrl,
    namespace: args.namespace,
    path: args.path,
    version: args.version,
  });
  return {
    queryKey: queryKeys.configVersion(args.namespace, args.path, args.version),
    queryFn: async (): Promise<GetVersionResponse> =>
      await apiFetch<GetVersionResponse>(url),
  };
}

export function useNamespaces(args: {
  baseUrl: string;
  limit?: number;
  enabled?: boolean;
  staleTime?: number;
}) {
  return useQuery({
    ...namespacesQueryOptions({ baseUrl: args.baseUrl, limit: args.limit }),
    enabled: args.enabled,
    staleTime: args.staleTime,
  });
}

export function useNamespaceBrowse(args: {
  baseUrl: string;
  namespace: string;
  prefix: string;
  enabled?: boolean;
  staleTime?: number;
}) {
  return useQuery({
    ...namespaceBrowseQueryOptions({
      baseUrl: args.baseUrl,
      namespace: args.namespace,
      prefix: args.prefix,
    }),
    enabled: args.enabled,
    staleTime: args.staleTime,
  });
}

export function useConfigLatest(args: {
  baseUrl: string;
  namespace: string;
  path: string;
  enabled?: boolean;
  staleTime?: number;
}) {
  return useQuery({
    ...configLatestQueryOptions({
      baseUrl: args.baseUrl,
      namespace: args.namespace,
      path: args.path,
    }),
    enabled: args.enabled,
    staleTime: args.staleTime,
  });
}

export async function invalidateNamespaceQueries(
  queryClient: QueryClient,
  namespace: string,
) {
  await queryClient.invalidateQueries({ queryKey: queryKeys.namespaces() });
  // Invalidate all browse prefixes under this namespace.
  await queryClient.invalidateQueries({
    queryKey: ["namespaceBrowse", namespace],
  });
}

export async function invalidateConfigQueries(
  queryClient: QueryClient,
  namespace: string,
  path: string,
) {
  await queryClient.invalidateQueries({
    queryKey: queryKeys.configLatest(namespace, path),
  });
  await queryClient.invalidateQueries({
    queryKey: queryKeys.configVersions(namespace, path),
  });
  // Invalidate all cached versions for this config.
  await queryClient.invalidateQueries({
    queryKey: ["configVersion", namespace, path],
  });
  await invalidateNamespaceQueries(queryClient, namespace);
}

