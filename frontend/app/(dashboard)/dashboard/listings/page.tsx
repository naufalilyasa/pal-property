import Link from "next/link";

import { requireUser } from "@/features/auth/server/require-user";
import { DashboardRefreshButton } from "@/features/listings/components/dashboard-refresh-button";
import { DashboardListingsTable } from "@/features/listings/components/dashboard-listings-table";
import { getSellerListingsPage } from "@/features/listings/server/get-seller-listings";

export default async function DashboardListingsPage() {
  await requireUser({ intent: "seller", returnTo: "/dashboard/listings" });
  const listingsPage = await getSellerListingsPage();

  return (
    <div className="space-y-6">
      <section className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-xs uppercase tracking-[0.24em] text-[var(--accent)]">Listings</p>
          <h1 className="mt-2 text-3xl font-semibold text-[var(--ink)]">Seller inventory</h1>
        </div>
        <div className="flex gap-3">
          <DashboardRefreshButton />
          <Link className="inline-flex items-center justify-center rounded-full bg-[var(--accent)] px-5 py-3 text-sm font-semibold text-white transition hover:opacity-90" href="/dashboard/listings/new">
            New listing
          </Link>
        </div>
      </section>

      {listingsPage.data.length === 0 ? (
        <section className="rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-8 text-sm leading-7 text-slate-900">
          No listings yet. Create your first property record to start the seller workflow.
        </section>
      ) : (
        <DashboardListingsTable listings={listingsPage.data} />
      )}
    </div>
  );
}
