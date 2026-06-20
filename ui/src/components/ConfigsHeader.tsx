"use client";

import Link from "next/link";
import { useState } from "react";

import { logout } from "@/lib/auth";
import type { AuthSession } from "@/lib/auth";

type ConfigsHeaderProps = {
  session: AuthSession;
};

export function ConfigsHeader({ session }: ConfigsHeaderProps) {
  const [busy, setBusy] = useState(false);

  async function handleLogout() {
    setBusy(true);
    try {
      await logout();
      window.location.href = "/configs";
    } finally {
      setBusy(false);
    }
  }

  const label =
    session.auth_enabled === false
      ? "Local dev"
      : session.email ?? session.actor_id ?? "Signed in";

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
        </div>
        {session.auth_enabled !== false ? (
          <div className="flex items-center gap-3 text-xs text-zinc-600 dark:text-zinc-400">
            <span className="hidden sm:inline">{label}</span>
            <button
              type="button"
              onClick={handleLogout}
              disabled={busy}
              className="rounded border border-black/[.08] px-2 py-1 text-xs hover:bg-zinc-100 disabled:opacity-60 dark:border-white/[.145] dark:hover:bg-zinc-900"
            >
              {busy ? "Signing out..." : "Logout"}
            </button>
          </div>
        ) : null}
      </div>
    </header>
  );
}
