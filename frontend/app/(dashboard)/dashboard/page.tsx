import Link from "next/link";

import { requireUser } from "@/features/auth/server/require-user";
import { DashboardRefreshButton } from "@/features/listings/components/dashboard-refresh-button";
import { getSellerListingsPage } from "@/features/listings/server/get-seller-listings";

export default async function DashboardOverviewPage() {
  await requireUser();
  const listingsPage = await getSellerListingsPage();

  return (
    <div className="space-y-6" data-testid="dashboard-shell">
      <section className="grid gap-4 md:grid-cols-3">
        <div className="rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-5">
          <p className="text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Total listings</p>
          <p className="mt-3 text-3xl font-semibold text-[var(--ink)]">{listingsPage.total}</p>
        </div>
        <div className="rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-5">
          <p className="text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Current page</p>
          <p className="mt-3 text-3xl font-semibold text-[var(--ink)]">{listingsPage.page}</p>
        </div>
        <div className="rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-5">
          <p className="text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Workflow</p>
          <p className="mt-3 text-sm leading-7 text-[var(--muted)]">Create, update, and curate listing images while the backend stays the auth and data authority.</p>
        </div>
      </section>

      <section className="flex flex-wrap gap-3">
        <DashboardRefreshButton />
        <Link className="inline-flex items-center justify-center rounded-full border border-[var(--line)] bg-[var(--panel)] px-5 py-3 text-sm font-semibold text-[var(--ink)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]" href="/dashboard/listings">
          Review listings table
        </Link>
        <Link className="inline-flex items-center justify-center rounded-full bg-[var(--accent)] px-5 py-3 text-sm font-semibold text-white transition hover:opacity-90" href="/dashboard/listings/new">
          Create listing
        </Link>
      </section>
    </div>
  );
}
