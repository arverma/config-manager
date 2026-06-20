export type AuthSession = {
  authenticated: boolean;
  auth_enabled?: boolean;
  email?: string;
  actor_type?: string;
  actor_id?: string;
};

export function loginRedirectPath(returnTo: string): string {
  const qp = new URLSearchParams();
  qp.set("returnTo", returnTo);
  return `/api/auth/login/google?${qp.toString()}`;
}

export function redirectToLogin(returnTo?: string): void {
  if (typeof window === "undefined") return;
  const path = returnTo ?? `${window.location.pathname}${window.location.search}`;
  window.location.href = loginRedirectPath(path);
}

export async function fetchSession(): Promise<AuthSession | null> {
  const res = await fetch("/api/auth/session", { cache: "no-store", credentials: "include" });
  if (res.status === 401) {
    return null;
  }
  if (!res.ok) {
    throw new Error(`session check failed: ${res.status}`);
  }
  return (await res.json()) as AuthSession;
}

export async function logout(): Promise<void> {
  await fetch("/api/auth/logout", {
    method: "POST",
    credentials: "include",
  });
}
