import Link from "next/link";

import { publicEnv } from "@/lib/env/public";

export default async function LoginPage({ searchParams }: { searchParams?: Promise<{ reason?: string }> }) {
  const resolvedSearchParams = searchParams ? await searchParams : {};
  const { reason } = resolvedSearchParams;
  const googleAuthUrl = `${publicEnv.NEXT_PUBLIC_API_BASE_URL.replace(/\/$/, "")}/auth/oauth/google`;

  return (
    <main className="min-h-screen px-6 py-10 sm:px-10 lg:px-12">
      <div className="mx-auto flex min-h-[calc(100vh-5rem)] w-full max-w-3xl flex-col justify-center rounded-[2rem] border border-white/60 bg-[var(--panel)] p-8 shadow-[0_30px_100px_rgba(15,23,42,0.18)] backdrop-blur sm:p-10">
        <p className="text-xs font-semibold uppercase tracking-[0.3em] text-[var(--accent)]">Authentication</p>
        <h1 className="mt-4 text-4xl font-semibold tracking-[-0.04em] text-[var(--ink)]">Continue to the seller workspace</h1>
        <p className="mt-4 text-sm leading-7 text-[var(--muted)]">
          Google OAuth stays in the Go backend. This frontend only starts the flow and reads the backend-owned cookie session.
        </p>

        {reason === "session-expired" ? (
          <div className="mt-6 rounded-[1.5rem] border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800" data-testid="auth-status-banner">
            Your session expired. Sign in again to continue managing listings.
          </div>
        ) : null}

        <div className="mt-8 flex flex-col gap-4 sm:flex-row">
          <a className="inline-flex items-center justify-center rounded-full bg-[var(--accent)] px-6 py-3 text-sm font-semibold text-white transition hover:opacity-90" data-testid="login-google-button" href={googleAuthUrl}>
            Continue with Google
          </a>
          <Link className="inline-flex items-center justify-center rounded-full border border-[var(--line)] bg-[var(--panel-strong)] px-6 py-3 text-sm font-semibold text-[var(--ink)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]" href="/">
            Back to home
          </Link>
        </div>
      </div>
    </main>
  );
}
