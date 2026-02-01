import type { ReactNode } from "react";

import { ConfigsHeader } from "@/components/ConfigsHeader";

export default function ConfigsLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-zinc-50 font-sans text-zinc-900 dark:bg-black dark:text-zinc-50">
      <ConfigsHeader />
      <main className="mx-auto w-full max-w-6xl px-6 py-8">{children}</main>
    </div>
  );
}

