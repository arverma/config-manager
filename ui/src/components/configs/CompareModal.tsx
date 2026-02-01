"use client";

import type { ReactCodeMirrorProps } from "@uiw/react-codemirror";

import { CodeEditor } from "@/components/shared/CodeEditor";

export type CompareVersionMeta = {
  version: number;
};

export function CompareModal(props: {
  error: string | null;
  loading: boolean;
  savingLeft: boolean;
  savingRight: boolean;

  versions: CompareVersionMeta[] | null;
  leftVersion: number | null;
  rightVersion: number | null;

  leftText: string;
  rightText: string;

  leftExtensions: NonNullable<ReactCodeMirrorProps["extensions"]>;
  rightExtensions: NonNullable<ReactCodeMirrorProps["extensions"]>;
  theme: ReactCodeMirrorProps["theme"];

  saveLeftDisabledNoChanges: boolean;
  saveRightDisabledNoChanges: boolean;

  onClose: () => void;
  onSwap: () => void;
  onSaveLeft: () => void;
  onSaveRight: () => void;

  onLeftTextChange: (v: string) => void;
  onRightTextChange: (v: string) => void;

  onLeftVersionChange: (v: number) => void | Promise<void>;
  onRightVersionChange: (v: number) => void | Promise<void>;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-6xl overflow-hidden rounded-xl border border-black/[.08] bg-white shadow-xl dark:border-white/[.145] dark:bg-black">
        <div className="flex items-center justify-between gap-4 border-b border-black/[.08] p-4 dark:border-white/[.145]">
          <div className="flex flex-col gap-1">
            <div className="text-sm font-medium">Compare versions</div>
            <div className="text-xs text-zinc-600 dark:text-zinc-400">
              Changes are highlighted (best-effort). Edit either side and save
              to create a new version.
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              className="h-9 rounded-lg border border-black/[.08] px-3 text-sm hover:bg-zinc-950/[.03] disabled:opacity-60 dark:border-white/[.145] dark:hover:bg-white/[.04]"
              onClick={props.onSwap}
              disabled={props.loading || props.savingLeft || props.savingRight}
            >
              Swap
            </button>
            <button
              type="button"
              className="h-9 rounded-lg border border-black/[.08] px-3 text-sm hover:bg-zinc-950/[.03] dark:border-white/[.145] dark:hover:bg-white/[.04]"
              onClick={props.onClose}
              disabled={props.savingLeft || props.savingRight}
            >
              Close
            </button>
          </div>
        </div>

        {props.error ? (
          <div className="border-b border-red-500/30 bg-red-500/5 p-3 text-sm text-red-600 dark:text-red-400">
            {props.error}
          </div>
        ) : null}

        <div className="grid grid-cols-1 gap-4 p-4 md:grid-cols-2">
          <div className="rounded-xl border border-black/[.08] bg-white p-3 dark:border-white/[.145] dark:bg-black">
            <div className="mb-2 flex items-center justify-between gap-3">
              <div className="flex items-center gap-2 text-xs text-zinc-600 dark:text-zinc-400">
                <span className="font-medium text-zinc-900 dark:text-zinc-50">
                  Left
                </span>
                <select
                  className="h-8 rounded-lg border border-black/[.08] bg-transparent px-2 text-xs dark:border-white/[.145]"
                  value={props.leftVersion ?? undefined}
                  onChange={(e) => props.onLeftVersionChange(Number(e.target.value))}
                  disabled={
                    props.loading ||
                    props.savingLeft ||
                    props.savingRight ||
                    props.versions === null
                  }
                >
                  {(props.versions ?? []).map((m) => (
                    <option key={`l-${m.version}`} value={m.version}>
                      v{m.version}
                    </option>
                  ))}
                </select>
              </div>
              <button
                type="button"
                className="h-9 cursor-pointer rounded-lg bg-zinc-900 px-3 text-sm font-medium text-zinc-50 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
                onClick={props.onSaveLeft}
                disabled={
                  props.loading ||
                  props.savingLeft ||
                  props.savingRight ||
                  props.saveLeftDisabledNoChanges
                }
              >
                {props.savingLeft
                  ? "Saving..."
                  : props.saveLeftDisabledNoChanges
                    ? "No changes"
                    : "Save new version"}
              </button>
            </div>
            <div className="overflow-hidden rounded-lg border border-black/[.08] dark:border-white/[.145]">
              <CodeEditor
                value={props.leftText}
                height="360px"
                extensions={props.leftExtensions}
                theme={props.theme}
                onChange={props.onLeftTextChange}
              />
            </div>
          </div>

          <div className="rounded-xl border border-black/[.08] bg-white p-3 dark:border-white/[.145] dark:bg-black">
            <div className="mb-2 flex items-center justify-between gap-3">
              <div className="flex items-center gap-2 text-xs text-zinc-600 dark:text-zinc-400">
                <span className="font-medium text-zinc-900 dark:text-zinc-50">
                  Right
                </span>
                <select
                  className="h-8 rounded-lg border border-black/[.08] bg-transparent px-2 text-xs dark:border-white/[.145]"
                  value={props.rightVersion ?? undefined}
                  onChange={(e) =>
                    props.onRightVersionChange(Number(e.target.value))
                  }
                  disabled={
                    props.loading ||
                    props.savingLeft ||
                    props.savingRight ||
                    props.versions === null
                  }
                >
                  {(props.versions ?? []).map((m) => (
                    <option key={`r-${m.version}`} value={m.version}>
                      v{m.version}
                    </option>
                  ))}
                </select>
              </div>
              <button
                type="button"
                className="h-9 cursor-pointer rounded-lg bg-zinc-900 px-3 text-sm font-medium text-zinc-50 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
                onClick={props.onSaveRight}
                disabled={
                  props.loading ||
                  props.savingLeft ||
                  props.savingRight ||
                  props.saveRightDisabledNoChanges
                }
              >
                {props.savingRight
                  ? "Saving..."
                  : props.saveRightDisabledNoChanges
                    ? "No changes"
                    : "Save new version"}
              </button>
            </div>
            <div className="overflow-hidden rounded-lg border border-black/[.08] dark:border-white/[.145]">
              <CodeEditor
                value={props.rightText}
                height="360px"
                extensions={props.rightExtensions}
                theme={props.theme}
                onChange={props.onRightTextChange}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

