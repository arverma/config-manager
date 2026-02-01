"use client";

import { useState } from "react";

import type { ConfigFormat } from "@/lib/configApi";

export function CreateConfigButton(props: {
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

