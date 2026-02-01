"use client";

import { useEffect, useMemo } from "react";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";

import { ApiErrorBanner } from "@/components/shared/ApiErrorBanner";
import { LoadingState } from "@/components/shared/LoadingState";
import { NamespaceBrowserContainer } from "@/components/namespaces/NamespaceBrowserContainer";
import { CreateConfigEditor } from "@/components/CreateConfigEditor";
import { ConfigEditor } from "@/components/ConfigEditor";
import { HttpError } from "@/lib/api/client";
import { configLatestQueryOptions, useNamespaceBrowse } from "@/lib/api/hooks";
import type { GetConfigResponse } from "@/lib/api/types";
import { getConfigApiBaseUrl } from "@/lib/configApi";

export function ConfigPageClient() {
  const router = useRouter();
  const params = useParams<{ namespace: string; path?: string[] }>();
  const searchParams = useSearchParams();

  const namespace = params.namespace;
  const pathStr = (params.path ?? []).join("/");
  const prefix = pathStr ? `${pathStr}/` : "";

  const createMode = searchParams.get("create") === "1";
  const initialFormat = searchParams.get("format") === "json" ? "json" : "yaml";

  const baseUrl = getConfigApiBaseUrl();

  const latestOptions = useMemo(() => {
    return configLatestQueryOptions({
      baseUrl,
      namespace,
      path: pathStr || "__noop__",
    });
  }, [baseUrl, namespace, pathStr]);

  const configQuery = useQuery({
    queryKey: pathStr ? latestOptions.queryKey : ["noop"],
    enabled: Boolean(pathStr),
    queryFn: latestOptions.queryFn as () => Promise<GetConfigResponse>,
  });

  const browseQuery = useNamespaceBrowse({
    baseUrl,
    namespace,
    prefix,
  });

  // Namespace doesn't exist: go back to all namespaces.
  useEffect(() => {
    if (browseQuery.error instanceof HttpError && browseQuery.error.status === 404) {
      router.replace("/configs");
    }
  }, [browseQuery.error, router]);

  // If we tried to open a config leaf, got 404, and browsing also yields 0 items,
  // treat it as a non-existent path and bounce to namespace root.
  useEffect(() => {
    if (!pathStr) return;
    if (createMode) return;
    if (!configQuery.error) return;
    if (!(configQuery.error instanceof HttpError) || configQuery.error.status !== 404) return;
    if (!browseQuery.data || !Array.isArray(browseQuery.data.items)) return;
    if (browseQuery.data.items.length !== 0) return;
    router.replace(`/configs/${encodeURIComponent(namespace)}`);
  }, [browseQuery.data, configQuery.error, createMode, namespace, pathStr, router]);

  // If pathStr exists and config resolves, show editor.
  if (pathStr && configQuery.isLoading) {
    return <LoadingState label="Loading config..." />;
  }
  if (pathStr && configQuery.data) {
    return (
      <div className="flex flex-col gap-6">
        <ConfigEditor
          baseUrl={baseUrl}
          namespace={namespace}
          path={pathStr}
          initial={configQuery.data}
        />
      </div>
    );
  }
  if (pathStr && configQuery.error && !(configQuery.error instanceof HttpError && configQuery.error.status === 404)) {
    return <ApiErrorBanner title="API error" error={configQuery.error} />;
  }

  // Config 404 + create mode => create editor.
  if (
    pathStr &&
    createMode &&
    configQuery.error instanceof HttpError &&
    configQuery.error.status === 404
  ) {
    return (
      <div className="flex flex-col gap-6">
        <CreateConfigEditor
          baseUrl={baseUrl}
          namespace={namespace}
          path={pathStr}
          initialFormat={initialFormat}
        />
      </div>
    );
  }

  // Otherwise treat as folder prefix and browse.
  return (
    <div className="flex flex-col gap-6">
      <NamespaceBrowserContainer
        baseUrl={baseUrl}
        namespace={namespace}
        prefix={prefix}
      />
    </div>
  );
}

