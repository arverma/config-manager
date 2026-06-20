"use client";

import type { ReactNode } from "react";

import { AuthGate } from "@/components/AuthGate";
import { ConfigsHeader } from "@/components/ConfigsHeader";

export default function ConfigsLayout({ children }: { children: ReactNode }) {
  return (
    <AuthGate>
      {(session) => (
        <div className="min-h-screen bg-zinc-50 font-sans text-zinc-900 dark:bg-black dark:text-zinc-50">
          <ConfigsHeader session={session} />
          <main className="mx-auto w-full max-w-6xl px-6 py-8">{children}</main>
        </div>
      )}
    </AuthGate>
  );
}
