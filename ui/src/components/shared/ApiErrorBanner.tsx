"use client";

import { HttpError } from "@/lib/api/client";

export function ApiErrorBanner(props: {
  title?: string;
  error: unknown;
  className?: string;
}) {
  const msg =
    props.error instanceof HttpError
      ? props.error.message
      : props.error instanceof Error
        ? props.error.message
        : "Unknown error";

  return (
    <div
      className={
        props.className ??
        "rounded-xl border border-red-500/30 bg-red-500/5 p-4 text-sm text-red-700 dark:text-red-300"
      }
    >
      {props.title ? (
        <div className="mb-1 font-medium">{props.title}</div>
      ) : null}
      <div>{msg}</div>
    </div>
  );
}

