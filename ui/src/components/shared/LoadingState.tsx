"use client";

export function LoadingState(props: { label?: string }) {
  return (
    <div className="rounded-xl border border-black/[.08] bg-white p-4 text-sm text-zinc-700 dark:border-white/[.145] dark:bg-black dark:text-zinc-300">
      {props.label ?? "Loading..."}
    </div>
  );
}

