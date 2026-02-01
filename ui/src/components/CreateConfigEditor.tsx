"use client";

import { json as jsonLang } from "@codemirror/lang-json";
import { yaml as yamlLang } from "@codemirror/lang-yaml";
import { oneDark } from "@codemirror/theme-one-dark";
import { useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import type { ConfigFormat } from "@/lib/configApi";
import { buildConfigUrl, getConfigApiBaseUrl } from "@/lib/configApi";
import { apiFetch, HttpError } from "@/lib/api/client";
import { invalidateConfigQueries } from "@/lib/api/hooks";
import { defaultConfigBody, prettify } from "@/lib/utils/prettify";
import { CodeEditor } from "@/components/shared/CodeEditor";

export function CreateConfigEditor(props: {
  baseUrl: string;
  namespace: string;
  path: string;
  initialFormat: ConfigFormat;
}) {
  const router = useRouter();
  const baseUrl = props.baseUrl || getConfigApiBaseUrl();
  const queryClient = useQueryClient();
  const namespaceHref = `/configs/${encodeURIComponent(props.namespace)}`;

  const [format, setFormat] = useState<ConfigFormat>(props.initialFormat);
  const [editorValue, setEditorValue] = useState<string>(() =>
    prettify(props.initialFormat, defaultConfigBody(props.initialFormat)),
  );

  const extensions = useMemo(() => {
    return [format === "json" ? jsonLang() : yamlLang()];
  }, [format]);

  const url = useMemo(() => {
    return buildConfigUrl({
      baseUrl,
      namespace: props.namespace,
      path: props.path,
    });
  }, [baseUrl, props.namespace, props.path]);

  const createMutation = useMutation({
    mutationFn: async () => {
      await apiFetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ format, body_raw: editorValue }),
      });
    },
    onSuccess: async () => {
      await invalidateConfigQueries(queryClient, props.namespace, props.path);
      router.replace(
        `/configs/${encodeURIComponent(props.namespace)}/${encodeURI(
          props.path,
        )}`,
      );
      router.refresh();
    },
  });

  const openExisting = () => {
    router.replace(
      `/configs/${encodeURIComponent(props.namespace)}/${encodeURI(
        props.path,
      )}`,
    );
    router.refresh();
  };

  return (
    <div className="flex flex-col gap-4">
      <header className="flex flex-col gap-1">
        <div className="text-sm text-zinc-600 dark:text-zinc-400">
          <Link
            href={namespaceHref}
            className="font-medium text-zinc-900 hover:underline dark:text-zinc-50"
          >
            {props.namespace}
          </Link>
          <span className="text-zinc-500">/</span>
          <code>{props.path}</code>
        </div>
        <div className="text-xs text-zinc-600 dark:text-zinc-400">
          creating ·{" "}
          <span className="inline-flex items-center gap-1">
            format=
            <select
              className="h-7 rounded-md border border-black/[.08] bg-transparent px-2 text-xs dark:border-white/[.145]"
              value={format}
              onChange={(e) => {
                const next = e.target.value as ConfigFormat;
                setFormat(next);
                setEditorValue(prettify(next, defaultConfigBody(next)));
              }}
              disabled={createMutation.isPending}
            >
              <option value="yaml">yaml</option>
              <option value="json">json</option>
            </select>
          </span>
        </div>
      </header>

      {createMutation.error ? (
        <div className="rounded-xl border border-red-500/30 bg-red-500/5 p-3 text-sm text-red-600 dark:text-red-400">
          <div>
            {createMutation.error instanceof HttpError &&
            createMutation.error.status === 409
              ? "Config already exists. Open the existing config instead."
              : createMutation.error instanceof Error
                ? createMutation.error.message
                : "Unknown error"}
          </div>
          {createMutation.error instanceof HttpError &&
          createMutation.error.status === 409 ? (
            <div className="mt-2">
              <button
                type="button"
                className="h-9 rounded-lg border border-black/[.08] px-3 text-sm hover:bg-zinc-950/[.03] dark:border-white/[.145] dark:hover:bg-white/[.04]"
                onClick={openExisting}
                disabled={createMutation.isPending}
              >
                Open existing
              </button>
            </div>
          ) : null}
        </div>
      ) : null}

      <section className="rounded-xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-black">
        <div className="mb-2 flex items-center justify-between gap-4">
          <div className="text-sm font-medium">Config</div>
          <button
            type="button"
            className="h-9 rounded-lg bg-zinc-900 px-3 text-sm font-medium text-zinc-50 hover:bg-zinc-800 disabled:opacity-60 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
            onClick={() => createMutation.mutate()}
            disabled={createMutation.isPending}
          >
            {createMutation.isPending ? "Creating..." : "Create config (v1)"}
          </button>
        </div>

        <div className="overflow-hidden rounded-lg border border-black/[.08] dark:border-white/[.145]">
          <CodeEditor
            value={editorValue}
            height="420px"
            extensions={extensions}
            theme={oneDark}
            onChange={(value) => setEditorValue(value)}
          />
        </div>
      </section>
    </div>
  );
}

