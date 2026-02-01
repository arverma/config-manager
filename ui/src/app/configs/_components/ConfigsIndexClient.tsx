"use client";

import Link from "next/link";

import { CreateNamespaceForm } from "@/components/CreateNamespaceForm";
import { ApiErrorBanner } from "@/components/shared/ApiErrorBanner";
import { LoadingState } from "@/components/shared/LoadingState";
import { useNamespaces } from "@/lib/api/hooks";

export function ConfigsIndexClient() {
  const namespacesQuery = useNamespaces({ limit: 500 });

  return (
    <div className="flex flex-col gap-8">
      <header className="flex flex-col gap-2">
        <h1 className="text-xl font-semibold tracking-tight">Namespaces</h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          Choose a namespace to browse configs.
        </p>
      </header>

      <section className="rounded-xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-black">
        <div className="mb-3 flex items-center justify-between gap-4">
          <div className="text-sm font-medium">All namespaces</div>
          <CreateNamespaceForm />
        </div>

        {namespacesQuery.isLoading ? (
          <LoadingState label="Loading namespaces..." />
        ) : namespacesQuery.error ? (
          <ApiErrorBanner title="API error" error={namespacesQuery.error} />
        ) : (
          <div className="divide-y divide-black/[.06] dark:divide-white/[.12]">
            {(namespacesQuery.data?.items ?? []).length === 0 ? (
              <div className="py-10 text-center text-sm text-zinc-600 dark:text-zinc-400">
                No namespaces yet.
              </div>
            ) : (
              (namespacesQuery.data?.items ?? []).map((ns) => (
                <Link
                  key={ns.id}
                  href={`/configs/${encodeURIComponent(ns.name)}`}
                  className="flex w-full items-center justify-between gap-4 rounded-lg px-4 py-3 hover:bg-zinc-950/[.03] dark:hover:bg-white/[.04]"
                >
                  <div className="flex flex-col">
                    <div className="text-sm font-medium">{ns.name}</div>
                  </div>
                  <div className="text-sm text-zinc-700 dark:text-zinc-300">
                    {ns.config_count} config{ns.config_count === 1 ? "" : "s"}
                  </div>
                </Link>
              ))
            )}
          </div>
        )}
      </section>
    </div>
  );
}

