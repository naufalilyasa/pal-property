import Link from "next/link";
import { TopNav } from "@/features/listings/components/top-nav";
import { Footer } from "@/features/listings/components/footer";
import { publicEnv } from "@/lib/env/public";

export default async function LoginPage({ searchParams }: { searchParams?: Promise<{ reason?: string }> }) {
  const resolvedSearchParams = searchParams ? await searchParams : {};
  const { reason } = resolvedSearchParams;
  const googleAuthUrl = `${publicEnv.NEXT_PUBLIC_API_BASE_URL.replace(/\/$/, "")}/auth/oauth/google`;

  return (
    <div className="flex min-h-screen flex-col bg-background font-sans text-foreground">
      <TopNav />
      
      <main className="flex flex-1 items-center justify-center p-6 sm:p-12">
        <div className="w-full max-w-md rounded-2xl border border-border bg-card p-8 shadow-lg sm:p-10">
          <div className="flex flex-col items-center text-center">
            <div className="mb-6 flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
              <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-primary"><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><polyline points="10 17 15 12 10 7"/><line x1="15" x2="3" y1="12" y2="12"/></svg>
            </div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Authentication</p>
            <h1 className="mt-4 text-3xl font-bold tracking-tight text-card-foreground">Sign in securely</h1>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
              Google OAuth stays in the Go backend. This frontend only starts the flow and reads the backend-owned cookie session.
            </p>
          </div>

          {reason === "session-expired" ? (
            <div className="mt-8 rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm font-medium text-destructive" data-testid="auth-status-banner">
              Your session expired. Sign in again to continue managing listings.
            </div>
          ) : null}

          <div className="mt-10 flex flex-col gap-3">
            <a className="inline-flex h-11 items-center justify-center gap-3 rounded-md bg-primary px-8 text-sm font-semibold text-primary-foreground shadow transition hover:bg-primary/90" data-testid="login-google-button" href={googleAuthUrl}>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="18" height="18"><path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/><path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/><path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/><path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/><path d="M1 1h22v22H1z" fill="none"/></svg>
              Continue with Google
            </a>
            <Link className="inline-flex h-11 items-center justify-center rounded-md border border-input bg-background px-8 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground" href="/">
              Back to home
            </Link>
          </div>
        </div>
      </main>

      <Footer />
    </div>
  );
}
