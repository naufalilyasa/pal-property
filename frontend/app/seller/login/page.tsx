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
      badge="Portal Agen / Seller"
      title="Manajemen Listing Anda"
      description="Masuk untuk menambah, mengedit, dan mengelola daftar properti yang Anda jual dengan mudah."
      primaryCtaLabel="Masuk dengan akun Google Agen"
      secondaryHref="/login"
      secondaryLabel="Bukan Agen? Masuk Publik"
      reason={reason}
      returnTo={returnTo}
      statusMessage="Sesi Anda telah berakhir. Silakan masuk kembali."
      accentClassName="bg-primary/15 ring-8 ring-primary/5"
    />
  );
}
