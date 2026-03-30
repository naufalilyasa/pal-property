import { AuthEntryShell } from "@/features/auth/components/auth-entry-shell";

export default async function LoginPage({
  searchParams,
}: {
  searchParams?: Promise<{ reason?: string; returnTo?: string }>;
}) {
  const resolvedSearchParams = searchParams ? await searchParams : {};
  const { reason, returnTo } = resolvedSearchParams;

  return (
    <AuthEntryShell
      intent="public"
      badge="General access"
      title="Sign in across FIND"
      description="Use the shared backend-owned Google OAuth entry to continue browsing, save momentum, and unlock account-linked features without moving auth into the frontend."
      primaryCtaLabel="Continue with Google"
      secondaryHref="/"
      secondaryLabel="Back to home"
      reason={reason}
      returnTo={returnTo}
      statusMessage="Your session expired. Sign in again to continue."
      highlights={[
        "Keep your account session in the backend while this page simply starts the OAuth flow.",
        "Return to the general experience for browsing, saved activity, and account continuity.",
        "Use the same auth-intent contract shared with the seller workspace route.",
      ]}
      accentClassName="bg-primary/10"
    />
  );
}
