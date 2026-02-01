"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";

import type { BrowseResponse } from "@/lib/api/types";
import { CreateConfigButton } from "@/components/namespaces/CreateConfigButton";

function stripTrailingSlash(s: string): string {
  return s.endsWith("/") ? s.slice(0, -1) : s;
}

function joinPath(a: string, b: string): string {
  if (!a) return b;
  if (!b) return a;
  return `${stripTrailingSlash(a)}/${stripTrailingSlash(b)}`;
}

export function NamespaceBrowserView(props: {
  baseUrl: string;
  namespace: string;
  prefix: string; // ends with / or empty

  browse: BrowseResponse;

  namespaceCount: number | null;
  loadingCount: boolean;

  deleteBusy: boolean;
  deleteError: string | null;
  canDeleteNamespace: boolean;
  onDeleteNamespace: () => void;
}) {
  const router = useRouter();
  const [deleteOpen, setDeleteOpen] = useState(false);

  const namespace = props.namespace;
  const prefix = props.prefix;

  const entries = useMemo(() => {
    return props.browse.items ?? [];
  }, [props.browse.items]);

  const breadcrumbs = useMemo(() => {
    const segments = stripTrailingSlash(prefix)
      .split("/")
      .filter(Boolean);
    const parts: { name: string; fullPath: string }[] = [];
    let acc = "";
    for (const seg of segments) {
      acc = acc ? `${acc}/${seg}` : seg;
      parts.push({ name: seg, fullPath: acc });
    }
    return parts;
  }, [prefix]);

  return (
    <div className="flex flex-col gap-4">
      <header className="flex flex-col gap-2">
        <div className="text-sm text-zinc-600 dark:text-zinc-400">
          Namespace <code>{namespace}</code>
        </div>
        <div className="flex flex-wrap items-center gap-2 text-sm">
          <Link
            href={`/configs/${encodeURIComponent(namespace)}`}
            className="font-medium hover:underline"
          >
            {namespace}
          </Link>
          {breadcrumbs.map((b) => (
            <span key={b.fullPath} className="flex items-center gap-2">
              <span className="text-zinc-500">/</span>
              <Link
                href={`/configs/${encodeURIComponent(namespace)}/${encodeURI(
                  b.fullPath,
                )}`}
                className="hover:underline"
              >
                {b.name}
              </Link>
            </span>
          ))}
        </div>
      </header>

      <section className="rounded-xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-black">
        <div className="mb-3 flex items-center justify-between gap-4">
          <div className="text-sm font-medium">Browse</div>
          <CreateConfigButton
            namespace={namespace}
            prefix={prefix}
            onCreated={(fullPath, format) => {
              router.push(
                `/configs/${encodeURIComponent(namespace)}/${encodeURI(
                  fullPath,
                )}?create=1&format=${encodeURIComponent(format)}`,
              );
            }}
          />
        </div>

        {entries.length === 0 ? (
          <div className="py-10 text-center text-sm text-zinc-600 dark:text-zinc-400">
            No configs under this path.
          </div>
        ) : (
          <div className="divide-y divide-black/[.06] dark:divide-white/[.12]">
            {entries.map((e) => {
              const isFolder = e.type === "folder";
              const fullPath = isFolder
                ? stripTrailingSlash(e.full_path)
                : e.full_path;
              const href = `/configs/${encodeURIComponent(
                namespace,
              )}/${encodeURI(fullPath)}`;
              return (
                <Link
                  key={`${e.type}:${e.full_path}`}
                  href={href}
                  className="flex w-full items-center justify-between gap-4 rounded-lg px-4 py-3 hover:bg-zinc-950/[.03] dark:hover:bg-white/[.04]"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-6 text-center text-zinc-500">
                      {isFolder ? "▸" : "•"}
                    </div>
                    <div className="flex flex-col">
                      <div className="text-sm font-medium">{e.name}</div>
                      {isFolder ? (
                        <div className="text-xs text-zinc-600 dark:text-zinc-400">
                          folder
                        </div>
                      ) : (
                        <div className="text-xs text-zinc-600 dark:text-zinc-400">
                          {e.format} · latest v{e.latest_version}
                        </div>
                      )}
                    </div>
                  </div>
                  <div className="text-xs text-zinc-600 dark:text-zinc-400">
                    {joinPath(namespace, fullPath)}
                  </div>
                </Link>
              );
            })}
          </div>
        )}
      </section>

      {prefix === "" ? (
        <section className="rounded-xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-black">
          <div className="flex items-start justify-between gap-4">
            <div className="flex flex-col gap-1">
              <div className="text-sm font-medium">Danger zone</div>
              <div className="text-xs text-zinc-600 dark:text-zinc-400">
                Delete this namespace (only allowed when it has 0 configs).
              </div>
              {props.loadingCount || props.namespaceCount === null ? (
                <div className="mt-1 text-xs text-zinc-700 dark:text-zinc-300">
                  Checking namespace usage…
                </div>
              ) : null}
            </div>

            <button
              type="button"
              className="h-9 rounded-lg border border-red-500/30 px-3 text-sm text-red-700 hover:bg-red-500/10 disabled:opacity-60 dark:text-red-300"
              disabled={
                props.deleteBusy ||
                props.loadingCount ||
                props.namespaceCount === null ||
                !props.canDeleteNamespace
              }
              onClick={() => setDeleteOpen((v) => !v)}
            >
              Delete namespace
            </button>
          </div>

          {props.deleteError ? (
            <div className="mt-3 text-xs text-red-700 dark:text-red-300">
              {props.deleteError}
            </div>
          ) : null}

          {deleteOpen ? (
            <div className="mt-4 rounded-lg border border-red-500/30 bg-red-500/5 p-3">
              <div className="mt-1 text-xs text-zinc-700 dark:text-zinc-300">
                This will permanently delete the namespace. This cannot be undone.
              </div>
              <div className="mt-3 flex items-center justify-end gap-2">
                <button
                  type="button"
                  className="h-9 rounded-lg border border-black/[.08] px-3 text-sm hover:bg-zinc-950/[.03] dark:border-white/[.145] dark:hover:bg-white/[.04]"
                  onClick={() => setDeleteOpen(false)}
                  disabled={props.deleteBusy}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="h-9 rounded-lg bg-red-600 px-3 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-60"
                  onClick={() => props.onDeleteNamespace()}
                  disabled={props.deleteBusy || !props.canDeleteNamespace}
                >
                  {props.deleteBusy ? "Deleting..." : "Confirm delete"}
                </button>
              </div>
            </div>
          ) : null}
        </section>
      ) : null}
    </div>
  );
}

