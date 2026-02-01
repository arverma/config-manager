"use client";

import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";

import { getConfigApiBaseUrl } from "@/lib/configApi";

export function CreateNamespaceForm() {
  const router = useRouter();
  const baseUrl = getConfigApiBaseUrl();

  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const nameIsValid = /^[a-z_]+$/.test(name.trim());

  const endpoint = useMemo(() => {
    return `${baseUrl}/namespaces`;
  }, [baseUrl]);

  const create = async () => {
    setSaving(true);
    setError(null);
    try {
      const res = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name }),
      });
      if (!res.ok) {
        const text = await res.text();
        setError(`API ${res.status}: ${text || res.statusText}`);
        return;
      }
      setOpen(false);
      setName("");
      router.refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setSaving(false);
    }
  };

  if (!open) {
    return (
      <button
        type="button"
        className="h-9 cursor-pointer rounded-lg bg-zinc-900 px-3 text-sm font-medium text-zinc-50 hover:bg-zinc-800 disabled:cursor-not-allowed dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
        onClick={() => setOpen(true)}
      >
        Create namespace
      </button>
    );
  }

  return (
    <div className="flex flex-col gap-2 rounded-lg border border-black/[.08] bg-white p-3 dark:border-white/[.145] dark:bg-black">
      <div className="text-sm font-medium">New namespace</div>
      <input
        className="h-9 rounded-lg border border-black/[.08] bg-transparent px-3 text-sm dark:border-white/[.145]"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="data_engg"
      />
      {error ? (
        <div className="text-xs text-red-600 dark:text-red-400">{error}</div>
      ) : (
        <div className="text-xs text-zinc-600 dark:text-zinc-400">
          Only <code>a-z</code> and <code>_</code> allowed.
        </div>
      )}
      {!error && name.trim().length > 0 && !nameIsValid ? (
        <div className="text-xs text-red-600 dark:text-red-400">
          Invalid name. Use lowercase letters and underscore only (pattern{" "}
          <code>^[a-z_]+$</code>).
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
          disabled={saving}
        >
          Cancel
        </button>
        <button
          type="button"
          className="h-9 cursor-pointer rounded-lg bg-zinc-900 px-3 text-sm font-medium text-zinc-50 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
          onClick={create}
          disabled={saving || name.trim().length === 0 || !nameIsValid}
        >
          {saving ? "Creating..." : "Create"}
        </button>
      </div>
    </div>
  );
}

