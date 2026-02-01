"use client";

import Link from "next/link";

export function ConfigsHeader() {
  return (
    <header className="border-b border-black/[.08] bg-white/70 backdrop-blur dark:border-white/[.145] dark:bg-black/40">
      <div className="mx-auto flex w-full max-w-6xl items-center justify-between gap-4 px-6 py-4">
        <div className="flex items-center gap-3">
          <Link
            href="/configs"
            className="text-sm font-semibold tracking-tight hover:opacity-80"
          >
            Config Manager
          </Link>
          <div className="hidden text-xs text-zinc-600 dark:text-zinc-400 md:block">
          </div>
        </div>
      </div>
    </header>
  );
}

