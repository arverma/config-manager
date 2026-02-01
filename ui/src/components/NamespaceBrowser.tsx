"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

import type { ConfigFormat } from "@/lib/configApi";
import { buildDeleteNamespaceUrl, buildNamespacesUrl } from "@/lib/configApi";

type BrowseEntry =
  | { type: "folder"; name: string; full_path: string }
  | {
      type: "config";
      name: string;
      full_path: string;
      format: ConfigFormat;
      latest_version: number;
    };

function stripTrailingSlash(s: string): string {
  return s.endsWith("/") ? s.slice(0, -1) : s;
}

function joinPath(a: string, b: string): string {
  if (!a) return b;
  if (!b) return a;
  return `${stripTrailingSlash(a)}/${stripTrailingSlash(b)}`;
}

export function NamespaceBrowser(props: {
  baseUrl: string;
  namespace: string;
  prefix: string; // ends with / or empty
  initial: { items: unknown[] } | null;
  status: number;
  statusText: string;
}) {
  const router = useRouter();

  const namespace = props.namespace;
  const prefix = props.prefix;

  const [namespaceCount, setNamespaceCount] = useState<number | null>(null);
  const [loadingCount, setLoadingCount] = useState(false);
  const [deleteBusy, setDeleteBusy] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const entries = useMemo(() => {
    const raw = props.initial?.items ?? [];
    return raw as BrowseEntry[];
  }, [props.initial]);

  // Fetch namespace config_count to decide if delete is allowed.
  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      if (prefix !== "") return; // only show delete on namespace root page
      setLoadingCount(true);
      setDeleteError(null);
      try {
        const url = buildNamespacesUrl({ baseUrl: props.baseUrl });
        // ask for a bigger page to avoid pagination issues in dev
        const res = await fetch(`${url}?limit=500`, { cache: "no-store" });
        if (!res.ok) return;
        const json = (await res.json()) as { items?: { name: string; config_count: number }[] };
        const match = (json.items ?? []).find((n) => n.name === namespace);
        if (!cancelled) setNamespaceCount(match ? match.config_count : 0);
      } catch {
        // ignore; UI will simply not show enabled delete
      } finally {
        if (!cancelled) setLoadingCount(false);
      }
    };
    void run();
    return () => {
      cancelled = true;
    };
  }, [namespace, prefix, props.baseUrl]);

  const canDeleteNamespace = namespaceCount === 0;

  const deleteNamespace = async () => {
    setDeleteError(null);
    if (!canDeleteNamespace) {
      setDeleteError("Namespace must be empty to delete.");
      return;
    }
    const ok = window.confirm(`Delete namespace '${namespace}'?\n\nThis cannot be undone.`);
    if (!ok) return;

    setDeleteBusy(true);
    try {
      const url = buildDeleteNamespaceUrl({ baseUrl: props.baseUrl, namespace });
      const res = await fetch(url, { method: "DELETE" });
      if (!res.ok) {
        const text = await res.text();
        setDeleteError(`API ${res.status}: ${text || res.statusText}`);
        return;
      }
      router.push(`/configs`);
      router.refresh();
    } catch (e) {
      setDeleteError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setDeleteBusy(false);
    }
  };

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
          <Link href={`/configs/${encodeURIComponent(namespace)}`} className="font-medium hover:underline">
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
          {props.status === 404 || !props.initial ? null : (
            <CreateConfigButton
              baseUrl={props.baseUrl}
              namespace={namespace}
              prefix={prefix}
              onCreated={(fullPath, format) => {
                router.push(
                  `/configs/${encodeURIComponent(namespace)}/${encodeURI(
                    fullPath,
                  )}?create=1&format=${encodeURIComponent(
                    format,
                  )}`,
                );
              }}
            />
          )}
        </div>

        {props.status === 404 ? (
          <div className="py-10 text-center text-sm text-zinc-600 dark:text-zinc-400">
            Namespace not found.
          </div>
        ) : !props.initial ? (
          <div className="py-10 text-center text-sm text-zinc-600 dark:text-zinc-400">
            API error: {props.status} {props.statusText}
          </div>
        ) : entries.length === 0 ? (
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
        <section className="rounded-xl border border-red-500/30 bg-red-500/5 p-4">
          <div className="text-sm font-medium text-red-700 dark:text-red-300">
            Danger zone
          </div>
          <div className="mt-1 text-xs text-zinc-700 dark:text-zinc-300">
            Delete this namespace (only allowed when it has 0 configs).
          </div>
          {deleteError ? (
            <div className="mt-2 text-xs text-red-700 dark:text-red-300">
              {deleteError}
            </div>
          ) : null}
          <div className="mt-3 flex items-center justify-between gap-3">
            <div className="text-xs text-zinc-700 dark:text-zinc-300">
              {loadingCount ? (
                <>Checking namespace usage…</>
              ) : namespaceCount === null ? (
                <>Checking namespace usage…</>
              ) : canDeleteNamespace ? (
                <>Namespace is empty.</>
              ) : (
                <>Namespace has {namespaceCount} config(s).</>
              )}
            </div>
            <button
              type="button"
              className="h-9 rounded-lg border border-red-500/30 px-3 text-sm text-red-700 hover:bg-red-500/10 disabled:opacity-60 dark:text-red-300"
              disabled={deleteBusy || loadingCount || namespaceCount === null || !canDeleteNamespace}
              onClick={deleteNamespace}
            >
              {deleteBusy ? "Deleting..." : "Delete namespace"}
            </button>
          </div>
        </section>
      ) : null}
    </div>
  );
}

function CreateConfigButton(props: {
  baseUrl: string;
  namespace: string;
  prefix: string; // ends with / or empty
  onCreated: (fullPath: string, format: ConfigFormat) => void;
}) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [format, setFormat] = useState<ConfigFormat>("yaml");
  const [error, setError] = useState<string | null>(null);

  const trimmedName = name.trim();
  const nameHasWhitespace = /\s/.test(trimmedName);
  const nameHasSlash = trimmedName.includes("/");
  const nameIsValid = trimmedName.length > 0 && !nameHasWhitespace && !nameHasSlash;

  const next = () => {
    setError(null);
    if (!nameIsValid) return;
    const fullPath = `${props.prefix}${trimmedName}`.replace(/\/+$/, "");
    setOpen(false);
    setName("");
    props.onCreated(fullPath, format);
  };

  if (!open) {
    return (
      <button
        type="button"
        className="h-9 cursor-pointer rounded-lg bg-zinc-900 px-3 text-sm font-medium text-zinc-50 hover:bg-zinc-800 disabled:cursor-not-allowed dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
        onClick={() => setOpen(true)}
      >
        Create config
      </button>
    );
  }

  return (
    <div className="flex w-full flex-col gap-2 rounded-lg border border-black/[.08] bg-white p-3 dark:border-white/[.145] dark:bg-black md:w-[520px]">
      <div className="text-sm font-medium">New config</div>

      <div className="grid grid-cols-1 gap-2 md:grid-cols-2">
        <label className="flex flex-col gap-1">
          <span className="text-xs text-zinc-600 dark:text-zinc-400">Name</span>
          <input
            className="h-9 rounded-lg border border-black/[.08] bg-transparent px-3 text-sm dark:border-white/[.145]"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="example"
          />
        </label>
        <label className="flex flex-col gap-1">
          <span className="text-xs text-zinc-600 dark:text-zinc-400">Format</span>
          <select
            className="h-9 rounded-lg border border-black/[.08] bg-transparent px-3 text-sm dark:border-white/[.145]"
            value={format}
            onChange={(e) => setFormat(e.target.value as ConfigFormat)}
          >
            <option value="yaml">yaml</option>
            <option value="json">json</option>
          </select>
        </label>
      </div>

      {error ? (
        <div className="text-xs text-red-600 dark:text-red-400">{error}</div>
      ) : (
        <div className="text-xs text-zinc-600 dark:text-zinc-400">
          Will be created at{" "}
          <code>
            /configs/{props.namespace}/{props.prefix}
            {trimmedName}
          </code>
        </div>
      )}
      {!error && trimmedName.length > 0 && nameHasWhitespace ? (
        <div className="text-xs text-red-600 dark:text-red-400">
          Config name must not contain whitespace.
        </div>
      ) : null}
      {!error && trimmedName.length > 0 && nameHasSlash ? (
        <div className="text-xs text-red-600 dark:text-red-400">
          Config name must not contain <code>/</code>.
        </div>
      ) : null}

      <div className="flex items-center justify-end gap-2">
        <button
          type="button"
          className="h-9 cursor-pointer rounded-lg border border-black/[.08] px-3 text-sm hover:bg-zinc-950/[.03] disabled:cursor-not-allowed dark:border-white/[.145] dark:hover:bg-white/[.04]"
          onClick={() => {
            setOpen(false);
            setName("");
            setError(null);
          }}
        >
          Cancel
        </button>
        <button
          type="button"
          className="h-9 cursor-pointer rounded-lg bg-zinc-900 px-3 text-sm font-medium text-zinc-50 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
          onClick={next}
          disabled={!nameIsValid}
        >
          Next
        </button>
      </div>
    </div>
  );
}

