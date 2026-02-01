"use client";

import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";

type ConfigFormat = "json" | "yaml";

function normalizePath(path: string): string {
  // UI mirrors REST URL: /configs/{namespace}/{path}
  // Keep it minimal: trim spaces and leading/trailing slashes.
  const trimmed = path.trim().replace(/^\/+/, "").replace(/\/+$/, "");
  return trimmed;
}

export function ConfigNavigator() {
  const router = useRouter();

  const [format, setFormat] = useState<ConfigFormat>("yaml");
  const [namespace, setNamespace] = useState("default");
  const [path, setPath] = useState("pipelines/example");

  const pathHasWhitespace = useMemo(() => /\s/.test(normalizePath(path)), [path]);

  const href = useMemo(() => {
    const ns = namespace.trim();
    const p = normalizePath(path);
    const qp = new URLSearchParams({ format });
    const suffix = qp.toString() ? `?${qp.toString()}` : "";
    return `/configs/${encodeURIComponent(ns)}/${p}${suffix}`;
  }, [format, namespace, path]);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        <div className="text-sm font-medium">Open config</div>
        <div className="text-xs text-zinc-600 dark:text-zinc-400">
          This URL shape matches the REST API: <code>/configs/&lt;namespace&gt;/&lt;path&gt;</code>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
        <label className="flex flex-col gap-1">
          <span className="text-sm">Format</span>
          <select
            className="h-10 rounded-lg border border-black/[.08] bg-transparent px-3 text-sm dark:border-white/[.145]"
            value={format}
            onChange={(e) => setFormat(e.target.value as ConfigFormat)}
          >
            <option value="yaml">yaml</option>
            <option value="json">json</option>
          </select>
        </label>

        <label className="flex flex-col gap-1">
          <span className="text-sm">Namespace</span>
          <input
            className="h-10 rounded-lg border border-black/[.08] bg-transparent px-3 text-sm dark:border-white/[.145]"
            value={namespace}
            onChange={(e) => setNamespace(e.target.value)}
            placeholder="payments"
          />
        </label>

        <label className="flex flex-col gap-1">
          <span className="text-sm">Path</span>
          <input
            className="h-10 rounded-lg border border-black/[.08] bg-transparent px-3 text-sm dark:border-white/[.145]"
            value={path}
            onChange={(e) => setPath(e.target.value)}
            placeholder="pipelines/orders_ingest"
          />
        </label>
      </div>

      <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
        <div className="text-xs text-zinc-600 dark:text-zinc-400">
          Target URL: <code className="break-all">{href}</code>
        </div>
        <button
          type="button"
          className="h-10 rounded-lg bg-zinc-900 px-4 text-sm font-medium text-zinc-50 hover:bg-zinc-800 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
          onClick={() => router.push(href)}
          disabled={pathHasWhitespace}
        >
          Open
        </button>
      </div>
      {pathHasWhitespace ? (
        <div className="text-xs text-red-600 dark:text-red-400">
          Path must not contain whitespace.
        </div>
      ) : null}
    </div>
  );
}

