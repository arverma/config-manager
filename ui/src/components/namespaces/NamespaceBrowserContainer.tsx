"use client";

import { useEffect } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

import { ApiErrorBanner } from "@/components/shared/ApiErrorBanner";
import { LoadingState } from "@/components/shared/LoadingState";
import { NamespaceBrowserView } from "./NamespaceBrowserView";
import { apiFetch, HttpError } from "@/lib/api/client";
import {
  invalidateNamespaceQueries,
  useNamespaceBrowse,
  useNamespaces,
} from "@/lib/api/hooks";
import { buildDeleteNamespaceUrl } from "@/lib/configApi";

export function NamespaceBrowserContainer(props: {
  baseUrl: string;
  namespace: string;
  prefix: string;
}) {
  const router = useRouter();
  const queryClient = useQueryClient();

  const browseQuery = useNamespaceBrowse({
    baseUrl: props.baseUrl,
    namespace: props.namespace,
    prefix: props.prefix,
  });

  // If namespace doesn't exist, bounce back to /configs.
  useEffect(() => {
    if (browseQuery.error instanceof HttpError && browseQuery.error.status === 404) {
      router.replace("/configs");
    }
  }, [browseQuery.error, router]);

  const namespacesQuery = useNamespaces({
    baseUrl: props.baseUrl,
    limit: 500,
    enabled: props.prefix === "",
    staleTime: 10_000,
  });

  const namespaceCount =
    namespacesQuery.data?.items?.find((n) => n.name === props.namespace)
      ?.config_count ?? null;
  const canDeleteNamespace = namespaceCount === 0;

  const deleteMutation = useMutation({
    mutationFn: async () => {
      if (!canDeleteNamespace) {
        throw new Error("Namespace must be empty to delete.");
      }
      const url = buildDeleteNamespaceUrl({
        baseUrl: props.baseUrl,
        namespace: props.namespace,
      });
      await apiFetch<void>(url, { method: "DELETE" });
    },
    onSuccess: async () => {
      await invalidateNamespaceQueries(queryClient, props.namespace);
      router.replace("/configs");
    },
  });

  if (browseQuery.isLoading) {
    return <LoadingState label="Loading folder..." />;
  }
  if (browseQuery.error) {
    if (browseQuery.error instanceof HttpError && browseQuery.error.status === 404) {
      return <LoadingState label="Redirecting..." />;
    }
    return <ApiErrorBanner title="API error" error={browseQuery.error} />;
  }

  if (!browseQuery.data) {
    return <LoadingState label="Loading folder..." />;
  }

  const deleteError =
    deleteMutation.error instanceof HttpError
      ? deleteMutation.error.message
      : deleteMutation.error instanceof Error
        ? deleteMutation.error.message
        : null;

  return (
    <NamespaceBrowserView
      baseUrl={props.baseUrl}
      namespace={props.namespace}
      prefix={props.prefix}
      browse={browseQuery.data}
      namespaceCount={namespaceCount}
      loadingCount={namespacesQuery.isFetching}
      deleteBusy={deleteMutation.isPending}
      deleteError={deleteError}
      canDeleteNamespace={canDeleteNamespace}
      onDeleteNamespace={() => {
        deleteMutation.mutate();
      }}
    />
  );
}

