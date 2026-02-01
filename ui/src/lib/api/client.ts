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

export async function apiFetch<T>(
  pathOrUrl: string,
  init?: RequestInit,
): Promise<T> {
  const baseUrl = getConfigApiBaseUrl();
  const url =
    pathOrUrl.startsWith("http://") || pathOrUrl.startsWith("https://")
      ? pathOrUrl
      : `${baseUrl}${pathOrUrl}`;

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

  // Some endpoints may return empty bodies (rare, but be defensive).
  const text = await res.text();
  if (!text) {
    return undefined as T;
  }
  return JSON.parse(text) as T;
}

