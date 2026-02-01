"use client";

import { json as jsonLang } from "@codemirror/lang-json";
import { yaml as yamlLang } from "@codemirror/lang-yaml";
import { oneDark } from "@codemirror/theme-one-dark";
import { useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import dynamic from "next/dynamic";

import {
  computeChangedLineNumbers,
  lineHighlightExtension,
} from "@/lib/cmDiffHighlight";
import {
  buildConfigPath,
  buildDeleteConfigPath,
  buildDeleteVersionPath,
  buildVersionPath,
} from "@/lib/configApi";
import { apiFetch, HttpError } from "@/lib/api/client";
import {
  configVersionQueryOptions,
  configVersionsQueryOptions,
  invalidateConfigQueries,
  invalidateNamespaceQueries,
} from "@/lib/api/hooks";
import { queryKeys } from "@/lib/api/keys";
import type {
  ConfigVersionMeta,
  GetConfigResponse,
  GetVersionResponse,
} from "@/lib/api/types";
import { prettify } from "@/lib/utils/prettify";
import { CodeEditor } from "@/components/shared/CodeEditor";
import { ApiErrorBanner } from "@/components/shared/ApiErrorBanner";
import { LoadingState } from "@/components/shared/LoadingState";

const CompareModal = dynamic(
  () => import("@/components/configs/CompareModal").then((m) => m.CompareModal),
  { ssr: false },
);

export function ConfigEditor(props: {
  namespace: string;
  path: string;
  initial?: unknown;
}) {
  const url = buildConfigPath({
    namespace: props.namespace,
    path: props.path,
  });

  const hasInitial = props.initial !== undefined && props.initial !== null;
  const latestQuery = useQuery({
    queryKey: queryKeys.configLatest(props.namespace, props.path),
    queryFn: async () => apiFetch<GetConfigResponse>(url),
    enabled: !hasInitial,
  });

  if (!hasInitial) {
    if (latestQuery.isLoading) {
      return <LoadingState label="Loading config..." />;
    }
    if (latestQuery.error) {
      return <ApiErrorBanner title="API error" error={latestQuery.error} />;
    }
    if (!latestQuery.data) {
      return <LoadingState label="Loading config..." />;
    }
    return (
      <ConfigEditorInner
        namespace={props.namespace}
        path={props.path}
        initial={latestQuery.data}
      />
    );
  }

  return (
    <ConfigEditorInner
      namespace={props.namespace}
      path={props.path}
      initial={props.initial as GetConfigResponse}
    />
  );
}

function ConfigEditorInner(props: {
  namespace: string;
  path: string;
  initial: GetConfigResponse;
}) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const namespaceHref = `/configs/${encodeURIComponent(
    props.namespace,
  )}`;

  const [data, setData] = useState<GetConfigResponse>(
    props.initial,
  );
  const [editorValue, setEditorValue] = useState<string>(() =>
    prettify(data.config.format, data.latest.body_raw),
  );
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [viewingVersion, setViewingVersion] = useState<number | null>(null);
  const [readOnly, setReadOnly] = useState(false);

  const [versions, setVersions] = useState<ConfigVersionMeta[] | null>(null);
  const [loadingVersions, setLoadingVersions] = useState(false);

  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteConfirmText, setDeleteConfirmText] = useState("");

  const [compareOpen, setCompareOpen] = useState(false);
  const [compareError, setCompareError] = useState<string | null>(null);
  const [loadingCompare, setLoadingCompare] = useState(false);
  const [savingCompareLeft, setSavingCompareLeft] = useState(false);
  const [savingCompareRight, setSavingCompareRight] = useState(false);

  const [leftVersion, setLeftVersion] = useState<number | null>(null);
  const [rightVersion, setRightVersion] = useState<number | null>(null);
  const [leftText, setLeftText] = useState<string>("");
  const [rightText, setRightText] = useState<string>("");

  const format = data.config.format;
  const extensions = useMemo(() => {
    return [format === "json" ? jsonLang() : yamlLang()];
  }, [format]);

  const refreshVersions = async (): Promise<ConfigVersionMeta[] | null> => {
    setLoadingVersions(true);
    setError(null);
    try {
      const resp = await queryClient.fetchQuery({
        ...configVersionsQueryOptions({
          namespace: props.namespace,
          path: props.path,
        }),
        staleTime: 5_000,
      });
      setVersions(resp.items);
      return resp.items;
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unknown error");
      return null;
    } finally {
      setLoadingVersions(false);
    }
  };

  const fetchVersionBodyRaw = async (version: number): Promise<string> => {
    const payload = await queryClient.fetchQuery({
      ...configVersionQueryOptions({
        namespace: props.namespace,
        path: props.path,
        version,
      }),
      staleTime: 30_000,
    });
    return prettify(format, payload.version.body_raw);
  };

  const openCompare = async (right: number) => {
    setCompareError(null);
    setCompareOpen(true);
    setLoadingCompare(true);
    try {
      const versionOptions = versions ?? (await refreshVersions()) ?? [];

      const lv = data.latest.version;
      const rv = right;

      setLeftVersion(lv);
      setRightVersion(rv);

      // Ensure selectors have stable options on first open.
      if (versions === null && versionOptions.length > 0) {
        setVersions(versionOptions);
      }

      const [lt, rt] = await Promise.all([
        fetchVersionBodyRaw(lv),
        fetchVersionBodyRaw(rv),
      ]);
      setLeftText(lt);
      setRightText(rt);
    } catch (e) {
      setCompareError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setLoadingCompare(false);
    }
  };

  const swapCompareSides = () => {
    const lv = leftVersion;
    const rv = rightVersion;
    const lt = leftText;
    const rt = rightText;
    setLeftVersion(rv);
    setRightVersion(lv);
    setLeftText(rt);
    setRightText(lt);
  };

  const saveCompare = async (side: "left" | "right") => {
    const body_raw = side === "left" ? leftText : rightText;
    const setBusy = side === "left" ? setSavingCompareLeft : setSavingCompareRight;

    const latestText = prettify(data.config.format, data.latest.body_raw);
    if (body_raw === latestText) {
      setCompareError("No changes to save.");
      return;
    }

    setBusy(true);
    setCompareError(null);
    try {
      const url = buildConfigPath({
        namespace: props.namespace,
        path: props.path,
      });
      const next = await apiFetch<GetConfigResponse>(url, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ body_raw }),
      });
      setData(next);
      setViewingVersion(null);
      setReadOnly(false);
      setEditorValue(prettify(next.config.format, next.latest.body_raw));
      setCompareOpen(false);
      await refreshVersions();
      await invalidateConfigQueries(queryClient, props.namespace, props.path);
    } catch (e) {
      if (e instanceof HttpError && e.status === 409 && e.code === "no_change") {
        setCompareError("No changes to save.");
      } else {
        setCompareError(e instanceof Error ? e.message : "Unknown error");
      }
    } finally {
      setBusy(false);
    }
  };

  const compareDiff = useMemo(() => {
    return computeChangedLineNumbers(leftText, rightText);
  }, [leftText, rightText]);

  const latestText = useMemo(() => {
    return prettify(data.config.format, data.latest.body_raw);
  }, [data.config.format, data.latest.body_raw]);

  const saveDisabledNoChanges =
    !readOnly && viewingVersion === null && editorValue === latestText;
  const saveCompareLeftDisabledNoChanges = leftText === latestText;
  const saveCompareRightDisabledNoChanges = rightText === latestText;

  const leftDiffExt = useMemo(() => {
    return lineHighlightExtension(compareDiff.left, "cm-diff-changed-left");
  }, [compareDiff.left]);

  const rightDiffExt = useMemo(() => {
    return lineHighlightExtension(compareDiff.right, "cm-diff-changed-right");
  }, [compareDiff.right]);

  const leftExtensions = useMemo(() => {
    return [...extensions, leftDiffExt];
  }, [extensions, leftDiffExt]);

  const rightExtensions = useMemo(() => {
    return [...extensions, rightDiffExt];
  }, [extensions, rightDiffExt]);

  const saveNewVersion = async () => {
    if (saveDisabledNoChanges) {
      setError("No changes to save.");
      return;
    }
    setSaving(true);
    setError(null);
    try {
      const url = buildConfigPath({
        namespace: props.namespace,
        path: props.path,
      });
      const next = await apiFetch<GetConfigResponse>(url, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ body_raw: editorValue }),
      });
      setData(next);
      setViewingVersion(null);
      setReadOnly(false);
      setEditorValue(prettify(next.config.format, next.latest.body_raw));
      await refreshVersions();
      await invalidateConfigQueries(queryClient, props.namespace, props.path);
    } catch (e) {
      if (e instanceof HttpError && e.status === 409 && e.code === "no_change") {
        setError("No changes to save.");
      } else {
        setError(e instanceof Error ? e.message : "Unknown error");
      }
    } finally {
      setSaving(false);
    }
  };

  const viewVersion = async (version: number, { editable }: { editable: boolean }) => {
    setSaving(true);
    setError(null);
    try {
      const url = buildVersionPath({
        namespace: props.namespace,
        path: props.path,
        version,
      });
      const payload = await apiFetch<GetVersionResponse>(url);
      setViewingVersion(payload.version.version);
      setEditorValue(prettify(data.config.format, payload.version.body_raw));
      setReadOnly(!editable);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setSaving(false);
    }
  };

  const openVersionForEdit = async (version: number) => {
    if (saving) return;

    const dirty = !readOnly && (viewingVersion !== null || editorValue !== latestText);
    if (dirty) {
      const ok = window.confirm("Discard unsaved changes?");
      if (!ok) return;
    }

    await viewVersion(version, { editable: true });
  };

  const backToLatest = async () => {
    setSaving(true);
    setError(null);
    try {
      const url = buildConfigPath({
        namespace: props.namespace,
        path: props.path,
      });
      const next = await apiFetch<GetConfigResponse>(url);
      setData(next);
      setViewingVersion(null);
      setReadOnly(false);
      setEditorValue(prettify(next.config.format, next.latest.body_raw));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setSaving(false);
    }
  };

  const deleteVersion = async (version: number) => {
    if (version === data.latest.version) {
      setError("Cannot delete the latest version.");
      return;
    }
    const ok = window.confirm(`Delete version v${version}? This cannot be undone.`);
    if (!ok) return;

    setSaving(true);
    setError(null);
    try {
      const url = buildDeleteVersionPath({
        namespace: props.namespace,
        path: props.path,
        version,
      });
      await apiFetch<void>(url, { method: "DELETE" });
      if (viewingVersion === version) {
        await backToLatest();
      }
      await refreshVersions();
      await invalidateConfigQueries(queryClient, props.namespace, props.path);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setSaving(false);
    }
  };

  const deleteConfig = async () => {
    const expected = `${props.namespace}/${props.path}`;
    if (deleteConfirmText.trim() !== expected) {
      setError(`Type '${expected}' to confirm deletion.`);
      return;
    }

    const ok = window.confirm(
      `Delete config '${expected}'?\n\nThis will hard-delete the config and its versions. This cannot be undone.`,
    );
    if (!ok) return;

    setSaving(true);
    setError(null);
    try {
      const url = buildDeleteConfigPath({
        namespace: props.namespace,
        path: props.path,
      });
      await apiFetch<void>(url, { method: "DELETE" });
      // Clear cached config data so we don't keep rendering stale data after a 404.
      queryClient.removeQueries({
        queryKey: queryKeys.configLatest(props.namespace, props.path),
      });
      queryClient.removeQueries({
        queryKey: queryKeys.configVersions(props.namespace, props.path),
      });
      queryClient.removeQueries({
        queryKey: ["configVersion", props.namespace, props.path],
      });
      await invalidateNamespaceQueries(queryClient, props.namespace);

      // Back to namespace browser
      router.replace(`/configs/${encodeURIComponent(props.namespace)}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col gap-4">
      <header className="flex flex-col gap-1">
        <div className="text-sm text-zinc-600 dark:text-zinc-400">
          <Link
            href={namespaceHref}
            className="font-medium text-zinc-900 hover:underline dark:text-zinc-50"
          >
            {data.config.namespace}
          </Link>
          <span className="text-zinc-500">/</span>
          <code>{data.config.path}</code>
        </div>
        <div className="text-xs text-zinc-600 dark:text-zinc-400">
          format=<code>{data.config.format}</code> · latest=v{data.latest.version}
          {viewingVersion !== null && viewingVersion !== data.latest.version ? (
            <>
              {" "}
              · viewing=v{viewingVersion}{" "}
              <button
                type="button"
                className="ml-2 underline hover:opacity-80"
                onClick={backToLatest}
                disabled={saving}
              >
                back to latest
              </button>
            </>
          ) : null}
        </div>
      </header>

      {error ? (
        <div className="rounded-xl border border-red-500/30 bg-red-500/5 p-3 text-sm text-red-600 dark:text-red-400">
          {error}
        </div>
      ) : null}

      <section className="rounded-xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-black">
        <div className="mb-2 flex items-center justify-between gap-4">
          <div className="text-sm font-medium">Config</div>
          <button
            type="button"
            className="h-9 cursor-pointer rounded-lg bg-zinc-900 px-3 text-sm font-medium text-zinc-50 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
            onClick={saveNewVersion}
            disabled={saving || readOnly || saveDisabledNoChanges}
          >
            {readOnly
              ? "Viewing"
              : saving
                ? "Saving..."
                : saveDisabledNoChanges
                  ? "No changes"
                  : "Save new version"}
          </button>
        </div>

        <div className="overflow-hidden rounded-lg border border-black/[.08] dark:border-white/[.145]">
          <CodeEditor
            value={editorValue}
            height="420px"
            extensions={extensions}
            theme={oneDark}
            onChange={(value) => setEditorValue(value)}
            editable={!readOnly}
          />
        </div>
      </section>

      <section className="rounded-xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-black">
        <div className="mb-3 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div className="text-sm font-medium">Versions</div>
          <div className="flex items-center justify-end gap-2">
            <button
              type="button"
              className="h-9 rounded-lg border border-red-500/30 px-3 text-sm text-red-600 hover:bg-red-500/5 disabled:opacity-60 dark:text-red-400"
              onClick={() => setDeleteOpen((v) => !v)}
              disabled={saving}
            >
              Delete config
            </button>
            <button
              type="button"
              className="h-9 rounded-lg border border-black/[.08] px-3 text-sm hover:bg-zinc-950/[.03] dark:border-white/[.145] dark:hover:bg-white/[.04]"
              onClick={refreshVersions}
              disabled={loadingVersions}
            >
              {loadingVersions ? "Loading..." : "Refresh"}
            </button>
          </div>
        </div>

        {deleteOpen ? (
          <div className="mb-4 rounded-lg border border-red-500/30 bg-red-500/5 p-3">
            <div className="text-sm font-medium text-red-700 dark:text-red-300">
              Danger zone
            </div>
            <div className="mt-1 text-xs text-zinc-700 dark:text-zinc-300">
              This will hard-delete the config and its versions. Type{" "}
              <code>{props.namespace}/{props.path}</code> to confirm.
            </div>
            <div className="mt-2 flex flex-col gap-2 md:flex-row md:items-center">
              <input
                className="h-9 flex-1 rounded-lg border border-red-500/30 bg-transparent px-3 text-sm"
                value={deleteConfirmText}
                onChange={(e) => setDeleteConfirmText(e.target.value)}
                placeholder={`${props.namespace}/${props.path}`}
              />
              <div className="flex items-center justify-end gap-2">
                <button
                  type="button"
                  className="h-9 rounded-lg border border-black/[.08] px-3 text-sm hover:bg-zinc-950/[.03] dark:border-white/[.145] dark:hover:bg-white/[.04]"
                  onClick={() => {
                    setDeleteOpen(false);
                    setDeleteConfirmText("");
                  }}
                  disabled={saving}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="h-9 rounded-lg bg-red-600 px-3 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-60"
                  onClick={deleteConfig}
                  disabled={saving || deleteConfirmText.trim() !== `${props.namespace}/${props.path}`}
                >
                  {saving ? "Deleting..." : "Confirm delete"}
                </button>
              </div>
            </div>
          </div>
        ) : null}

        {versions === null ? (
          <div className="text-sm text-zinc-600 dark:text-zinc-400">
            Click refresh to load versions.
          </div>
        ) : versions.length === 0 ? (
          <div className="text-sm text-zinc-600 dark:text-zinc-400">
            No versions.
          </div>
        ) : (
          <div className="divide-y divide-black/[.06] dark:divide-white/[.12]">
            {versions.map((v) => {
              const isLatest = v.version === data.latest.version;
              return (
                <div
                  key={v.id}
                  className="flex items-center justify-between gap-4 rounded-lg px-3 py-3 hover:bg-zinc-950/[.03] dark:hover:bg-white/[.04]"
                  role="button"
                  tabIndex={0}
                  aria-disabled={saving}
                  onClick={() => {
                    if (saving) return;
                    if (isLatest) {
                      if (viewingVersion !== null || readOnly) {
                        void backToLatest();
                      }
                      return;
                    }
                    void openVersionForEdit(v.version);
                  }}
                  onKeyDown={(e) => {
                    if (e.key !== "Enter" && e.key !== " ") return;
                    e.preventDefault();
                    if (saving) return;
                    if (isLatest) {
                      if (viewingVersion !== null || readOnly) {
                        void backToLatest();
                      }
                      return;
                    }
                    void openVersionForEdit(v.version);
                  }}
                >
                  <div className="flex flex-col">
                    <div className="text-sm font-medium">
                      v{v.version}{" "}
                      {isLatest ? (
                        <span className="ml-2 rounded-full bg-zinc-900 px-2 py-0.5 text-xs text-zinc-50 dark:bg-zinc-50 dark:text-zinc-900">
                          latest
                        </span>
                      ) : null}
                    </div>
                    <div className="text-xs text-zinc-600 dark:text-zinc-400">
                      {new Date(v.created_at).toLocaleString()}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      className="h-9 rounded-lg border border-black/[.08] px-3 text-sm hover:bg-zinc-950/[.03] disabled:opacity-60 dark:border-white/[.145] dark:hover:bg-white/[.04]"
                      disabled={saving || isLatest}
                      onClick={(e) => {
                        e.stopPropagation();
                        void openCompare(v.version);
                      }}
                      title={
                        isLatest ? "Already the latest" : "Compare this version to latest"
                      }
                    >
                      Compare to latest
                    </button>
                    <button
                      type="button"
                      className="h-9 rounded-lg border border-red-500/30 px-3 text-sm text-red-600 hover:bg-red-500/5 disabled:opacity-60 dark:text-red-400"
                      disabled={saving || isLatest}
                      onClick={(e) => {
                        e.stopPropagation();
                        void deleteVersion(v.version);
                      }}
                    >
                      Delete
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>

      {compareOpen ? (
        <CompareModal
          error={compareError}
          loading={loadingCompare}
          savingLeft={savingCompareLeft}
          savingRight={savingCompareRight}
          versions={versions}
          leftVersion={leftVersion}
          rightVersion={rightVersion}
          leftText={leftText}
          rightText={rightText}
          leftExtensions={leftExtensions}
          rightExtensions={rightExtensions}
          theme={oneDark}
          saveLeftDisabledNoChanges={saveCompareLeftDisabledNoChanges}
          saveRightDisabledNoChanges={saveCompareRightDisabledNoChanges}
          onClose={() => setCompareOpen(false)}
          onSwap={swapCompareSides}
          onSaveLeft={() => saveCompare("left")}
          onSaveRight={() => saveCompare("right")}
          onLeftTextChange={(v) => setLeftText(v)}
          onRightTextChange={(v) => setRightText(v)}
          onLeftVersionChange={async (v) => {
            setLeftVersion(v);
            setLoadingCompare(true);
            setCompareError(null);
            try {
              const t = await fetchVersionBodyRaw(v);
              setLeftText(t);
            } catch (err) {
              setCompareError(err instanceof Error ? err.message : "Unknown error");
            } finally {
              setLoadingCompare(false);
            }
          }}
          onRightVersionChange={async (v) => {
            setRightVersion(v);
            setLoadingCompare(true);
            setCompareError(null);
            try {
              const t = await fetchVersionBodyRaw(v);
              setRightText(t);
            } catch (err) {
              setCompareError(err instanceof Error ? err.message : "Unknown error");
            } finally {
              setLoadingCompare(false);
            }
          }}
        />
      ) : null}
    </div>
  );
}
