"use client";

import type { ReactNode } from "react";
import { useEffect, useState } from "react";

import { fetchSession, redirectToLogin, type AuthSession } from "@/lib/auth";
import { LoadingState } from "@/components/shared/LoadingState";

type AuthGateProps = {
  children: (session: AuthSession) => ReactNode;
};

export function AuthGate({ children }: AuthGateProps) {
  const [session, setSession] = useState<AuthSession | null | undefined>(undefined);

  useEffect(() => {
    let cancelled = false;

    fetchSession()
      .then((value) => {
        if (cancelled) return;
        if (!value?.authenticated) {
          redirectToLogin();
          return;
        }
        if (value.auth_enabled === false) {
          setSession({ authenticated: true, auth_enabled: false });
          return;
        }
        setSession(value);
      })
      .catch(() => {
        if (!cancelled) {
          setSession(null);
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  if (session === undefined) {
    return <LoadingState label="Checking session..." />;
  }

  if (!session) {
    return <LoadingState label="Redirecting to sign in..." />;
  }

  return <>{children(session)}</>;
}
