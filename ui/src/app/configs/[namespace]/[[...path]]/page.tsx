import { ConfigEditor } from "@/components/ConfigEditor";
import { CreateConfigEditor } from "@/components/CreateConfigEditor";
import { NamespaceBrowser } from "@/components/NamespaceBrowser";
import { redirect } from "next/navigation";
import {
  buildConfigUrl,
  buildNamespaceBrowseUrl,
  getConfigApiBaseUrl,
} from "@/lib/configApi";

type PageProps = {
  params: Promise<{
    namespace: string;
    path?: string[];
  }>;
  searchParams: Promise<{
    create?: string;
    format?: string;
  }>;
};

export default async function ConfigPage(props: PageProps) {
  const params = await props.params;
  const searchParams = await props.searchParams;

  const namespace = params.namespace;
  const pathStr = (params.path ?? []).join("/");
  const createMode = searchParams.create === "1";
  const initialFormat = searchParams.format === "json" ? "json" : "yaml";

  const baseUrl = getConfigApiBaseUrl();

  const prefix = pathStr ? `${pathStr}/` : "";
  const key = `${namespace}:${pathStr}`;

  let configWas404 = false;

  // If there's a path, first try to treat it as a config leaf.
  if (pathStr) {
    const configUrl = buildConfigUrl({ baseUrl, namespace, path: pathStr });
    const res = await fetch(configUrl, { cache: "no-store" });
    if (res.ok) {
      const data = (await res.json()) as unknown;
      return (
        <div className="flex flex-col gap-6">
          <ConfigEditor
            key={key}
            baseUrl={baseUrl}
            namespace={namespace}
            path={pathStr}
            initial={data}
          />
        </div>
      );
    }
    if (res.status !== 404) {
      return (
        <div className="rounded-xl border border-red-500/30 bg-red-500/5 p-4 text-sm text-red-600 dark:text-red-400">
          API error: {res.status} {res.statusText}
        </div>
      );
    }
    configWas404 = true;
    if (createMode) {
      return (
        <div className="flex flex-col gap-6">
          <CreateConfigEditor
            key={key}
            baseUrl={baseUrl}
            namespace={namespace}
            path={pathStr}
            initialFormat={initialFormat}
          />
        </div>
      );
    }
    // else: treat as folder prefix and browse.
  }

  const browseUrl = buildNamespaceBrowseUrl({
    baseUrl,
    namespace,
    prefix,
  });
  const browseRes = await fetch(browseUrl, { cache: "no-store" });
  const browseData = (await browseRes.json().catch(() => null)) as
    | { items: unknown[] }
    | null;

  // Namespace doesn't exist: go back to all namespaces.
  if (browseRes.status === 404) {
    redirect(`/configs`);
  }

  // If we were trying to open a config and it 404'd, but there is no folder
  // content under that prefix either, treat it as a non-existent path and
  // bounce to namespace root.
  if (
    pathStr &&
    configWas404 &&
    browseRes.ok &&
    Array.isArray(browseData?.items) &&
    browseData.items.length === 0
  ) {
    redirect(`/configs/${encodeURIComponent(namespace)}`);
  }

  return (
    <div className="flex flex-col gap-6">
      <NamespaceBrowser
        baseUrl={baseUrl}
        namespace={namespace}
        prefix={prefix}
        initial={browseData}
        status={browseRes.status}
        statusText={browseRes.statusText}
      />
    </div>
  );
}

