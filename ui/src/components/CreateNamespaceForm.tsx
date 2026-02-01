"use client";

import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { getConfigApiBaseUrl } from "@/lib/configApi";
import { apiFetch, HttpError } from "@/lib/api/client";
import { invalidateNamespaceQueries } from "@/lib/api/hooks";

export function CreateNamespaceForm() {
  const router = useRouter();
  const baseUrl = getConfigApiBaseUrl();
  const queryClient = useQueryClient();

  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");

  const nameIsValid = /^[a-z_]+$/.test(name.trim());

  const endpoint = useMemo(() => {
    return `${baseUrl}/namespaces`;
  }, [baseUrl]);

  const createMutation = useMutation({
    mutationFn: async () => {
      await apiFetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name }),
      });
    },
    onSuccess: async () => {
      const created = name.trim();
      setOpen(false);
      setName("");
      await invalidateNamespaceQueries(queryClient, created);
      router.refresh();
    },
  });

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
      {createMutation.error ? (
        <div className="text-xs text-red-600 dark:text-red-400">
          {createMutation.error instanceof HttpError
            ? createMutation.error.message
            : createMutation.error instanceof Error
              ? createMutation.error.message
              : "Unknown error"}
        </div>
      ) : (
        <div className="text-xs text-zinc-600 dark:text-zinc-400">
          Only <code>a-z</code> and <code>_</code> allowed.
        </div>
      )}
      {!createMutation.error && name.trim().length > 0 && !nameIsValid ? (
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
          }}
          disabled={createMutation.isPending}
        >
          Cancel
        </button>
        <button
          type="button"
          className="h-9 cursor-pointer rounded-lg bg-zinc-900 px-3 text-sm font-medium text-zinc-50 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
          onClick={() => createMutation.mutate()}
          disabled={
            createMutation.isPending || name.trim().length === 0 || !nameIsValid
          }
        >
          {createMutation.isPending ? "Creating..." : "Create"}
        </button>
      </div>
    </div>
  );
}

