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
      badge="Akses Publik"
      title="Selamat Datang di Pal Property"
      description="Masuk untuk dapatkan properti impian Anda dengan lebih mudah dan cepat."
      primaryCtaLabel="Lanjutkan dengan Google"
      secondaryHref="/"
      secondaryLabel="Kembali ke Beranda"
      reason={reason}
      returnTo={returnTo}
      statusMessage="Sesi Anda telah berakhir. Silakan masuk kembali."
      accentClassName="bg-primary/10"
    />
  );
}
