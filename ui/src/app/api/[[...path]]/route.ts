/**
 * API proxy: forwards /api/* to CONFIG_API_BASE_URL at request time.
 * Optional CONFIG_API_UPSTREAM_TIMEOUT_MS (default 30s). Forwards cookie, content-type, etc.;
 * request body uses duplex: "half". On timeout returns 504; on other upstream errors 502.
 */
import { NextRequest } from "next/server";

export const dynamic = "force-dynamic";

const UPSTREAM_TIMEOUT_MS = (() => {
  const v = process.env.CONFIG_API_UPSTREAM_TIMEOUT_MS;
  if (v == null || v === "") return 30_000;
  const n = parseInt(v, 10);
  return Number.isNaN(n) ? 30_000 : n;
})();

const upstreamBase =
  process.env.CONFIG_API_BASE_URL ||
  process.env.NEXT_PUBLIC_CONFIG_API_BASE_URL ||
  "http://localhost:8080";

function buildUpstreamUrl(request: NextRequest): string {
  const url = new URL(request.url);
  const pathAfterApi = url.pathname.replace(/^\/api\/?/, "") || "";
  const base = upstreamBase.replace(/\/$/, "");
  const pathPart = pathAfterApi ? `/${pathAfterApi}` : "";
  const search = url.search ? url.search : "";
  return `${base}${pathPart}${search}`;
}

const forwardHeaders = [
  "content-type",
  "accept",
  "accept-encoding",
  "authorization",
  "cookie",
];

function proxyRequest(request: NextRequest): Promise<Response> {
  const upstreamUrl = buildUpstreamUrl(request);
  const headers = new Headers();
  forwardHeaders.forEach((name) => {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  });

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), UPSTREAM_TIMEOUT_MS);

  const init: RequestInit = {
    method: request.method,
    headers,
    cache: "no-store",
    signal: controller.signal,
  };
  if (request.method !== "GET" && request.method !== "HEAD" && request.body) {
    init.body = request.body;
    (init as RequestInit & { duplex?: string }).duplex = "half";
  }

  return fetch(upstreamUrl, init)
    .then((res) => {
      clearTimeout(timeoutId);
      return res;
    })
    .catch((err) => {
      clearTimeout(timeoutId);
      const isTimeout = err?.name === "AbortError";
      console.error(
        "[api proxy] upstream fetch failed:",
        upstreamUrl,
        isTimeout ? "timeout" : err,
      );
      return new Response(
        JSON.stringify({
          code: isTimeout ? "UPSTREAM_TIMEOUT" : "UPSTREAM_ERROR",
          message: isTimeout
            ? "Upstream did not respond in time"
            : err instanceof Error
              ? err.message
              : "Upstream request failed",
        }),
        {
          status: isTimeout ? 504 : 502,
          headers: { "Content-Type": "application/json" },
        },
      );
    });
}

export function GET(request: NextRequest) {
  return proxyRequest(request);
}
export function POST(request: NextRequest) {
  return proxyRequest(request);
}
export function PUT(request: NextRequest) {
  return proxyRequest(request);
}
export function PATCH(request: NextRequest) {
  return proxyRequest(request);
}
export function DELETE(request: NextRequest) {
  return proxyRequest(request);
}
export function HEAD(request: NextRequest) {
  return proxyRequest(request);
}
export function OPTIONS(request: NextRequest) {
  return proxyRequest(request);
}
