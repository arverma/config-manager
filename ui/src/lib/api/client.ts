import { getConfigApiBaseUrl } from "@/lib/configApi";

export type ApiError = {
  code?: string;
  message?: string;
  details?: Record<string, unknown>;
};

export class HttpError extends Error {
  status: number;
  code?: string;
  details?: Record<string, unknown>;

  constructor(args: {
    status: number;
    message: string;
    code?: string;
    details?: Record<string, unknown>;
  }) {
    super(args.message);
    this.name = "HttpError";
    this.status = args.status;
    this.code = args.code;
    this.details = args.details;
  }
}

async function parseApiError(res: Response): Promise<ApiError | null> {
  const text = await res.text().catch(() => "");
  if (!text) return null;
  try {
    return JSON.parse(text) as ApiError;
  } catch {
    return { message: text };
  }
}

function safeParseJSON<T>(text: string): T {
  try {
    return JSON.parse(text) as T;
  } catch {
    throw new Error("Invalid JSON response");
  }
}

export async function apiFetch<T>(
  pathOrUrl: string,
  init?: RequestInit,
): Promise<T> {
  const baseUrl = getConfigApiBaseUrl();
  const isAbsHttp =
    pathOrUrl.startsWith("http://") || pathOrUrl.startsWith("https://");
  if (isAbsHttp) {
    return await apiFetchAbsolute<T>(pathOrUrl, init);
  }

  if (!pathOrUrl.startsWith("/")) {
    throw new Error("apiFetch expects a path starting with '/'");
  }

  // Guard: pass API path (e.g. /namespaces), not /api/namespaces.
  if (
    process.env.NODE_ENV !== "production" &&
    baseUrl.startsWith("/") &&
    (pathOrUrl === baseUrl || pathOrUrl.startsWith(`${baseUrl}/`))
  ) {
    throw new Error(
      `apiFetch path must not include baseUrl (${baseUrl}). Pass '/namespaces' not '${pathOrUrl}'.`,
    );
  }

  const url = joinBaseAndPath(baseUrl, pathOrUrl);

  const res = await fetch(url, {
    cache: "no-store",
    ...init,
    headers: {
      ...(init?.headers ?? {}),
    },
  });

  if (!res.ok) {
    const parsed = await parseApiError(res);
    throw new HttpError({
      status: res.status,
      code: parsed?.code,
      message: parsed?.message || `API ${res.status} ${res.statusText}`,
      details:
        parsed?.details && typeof parsed.details === "object"
          ? parsed.details
          : undefined,
    });
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const text = await res.text();
  if (!text) {
    return undefined as T;
  }
  return safeParseJSON<T>(text);
}

function joinBaseAndPath(baseUrl: string, path: string): string {
  // baseUrl can be absolute (http://...) or relative (/api).
  const b = baseUrl.endsWith("/") ? baseUrl.slice(0, -1) : baseUrl;
  return b ? `${b}${path}` : path;
}

async function apiFetchAbsolute<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    cache: "no-store",
    ...init,
    headers: {
      ...(init?.headers ?? {}),
    },
  });

  if (!res.ok) {
    const parsed = await parseApiError(res);
    throw new HttpError({
      status: res.status,
      code: parsed?.code,
      message: parsed?.message || `API ${res.status} ${res.statusText}`,
      details:
        parsed?.details && typeof parsed.details === "object"
          ? parsed.details
          : undefined,
    });
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const text = await res.text();
  if (!text) {
    return undefined as T;
  }
  return safeParseJSON<T>(text);
}

