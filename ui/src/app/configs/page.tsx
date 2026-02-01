import Link from "next/link";

import { CreateNamespaceForm } from "@/components/CreateNamespaceForm";
import { buildNamespacesUrl, getConfigApiBaseUrl } from "@/lib/configApi";

type NamespaceWithCount = {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
  config_count: number;
};

export default async function ConfigsIndexPage() {
  const baseUrl = getConfigApiBaseUrl();

  const url = buildNamespacesUrl({ baseUrl });
  const res = await fetch(url, { cache: "no-store" });
  const data = (await res.json().catch(() => null)) as
    | { items: NamespaceWithCount[] }
    | null;

  const items = data?.items ?? [];

  return (
    <div className="flex flex-col gap-8">
      <header className="flex flex-col gap-2">
        <h1 className="text-xl font-semibold tracking-tight">Namespaces</h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          Choose a namespace to browse configs.
        </p>
      </header>

      <section className="rounded-xl border border-black/[.08] bg-white p-4 dark:border-white/[.145] dark:bg-black">
        <div className="mb-3 flex items-center justify-between gap-4">
          <div className="text-sm font-medium">All namespaces</div>
          <CreateNamespaceForm />
        </div>

        {res.ok ? null : (
          <div className="mb-3 rounded-lg border border-red-500/30 bg-red-500/5 p-3 text-sm text-red-600 dark:text-red-400">
            API error: {res.status} {res.statusText}
          </div>
        )}

        <div className="divide-y divide-black/[.06] dark:divide-white/[.12]">
          {items.length === 0 ? (
            <div className="py-10 text-center text-sm text-zinc-600 dark:text-zinc-400">
              No namespaces yet.
            </div>
          ) : (
            items.map((ns) => (
              <Link
                key={ns.id}
                href={`/configs/${encodeURIComponent(ns.name)}`}
                className="flex w-full items-center justify-between gap-4 rounded-lg px-4 py-3 hover:bg-zinc-950/[.03] dark:hover:bg-white/[.04]"
              >
                <div className="flex flex-col">
                  <div className="text-sm font-medium">{ns.name}</div>
                </div>
                <div className="text-sm text-zinc-700 dark:text-zinc-300">
                  {ns.config_count} config{ns.config_count === 1 ? "" : "s"}
                </div>
              </Link>
            ))
          )}
        </div>
      </section>
    </div>
  );
}

