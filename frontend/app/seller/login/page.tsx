import { AuthEntryShell } from "@/features/auth/components/auth-entry-shell";

export default async function SellerLoginPage({
  searchParams,
}: {
  searchParams?: Promise<{ reason?: string; returnTo?: string }>;
}) {
  const resolvedSearchParams = searchParams ? await searchParams : {};
  const { reason, returnTo } = resolvedSearchParams;

  return (
    <AuthEntryShell
      intent="seller"
      badge="Seller workspace"
      title="Access your listing desk"
      description="Start the same backend-owned Google OAuth flow, then return to the seller workspace to manage drafts, images, and publishing tasks."
      primaryCtaLabel="Continue with Google for sellers"
      secondaryHref="/login"
      secondaryLabel="Need the general login page?"
      reason={reason}
      returnTo={returnTo}
      statusMessage="Your seller session expired. Sign in again to continue managing listings."
      highlights={[
        "Return to the seller workspace intent without changing the backend OAuth entrypoint.",
        "Keep listing creation, edits, and image management in one secure place.",
        "Use the same backend-owned cookie session once Google sign-in completes.",
      ]}
      accentClassName="bg-primary/15 ring-8 ring-primary/5"
    />
  );
}
