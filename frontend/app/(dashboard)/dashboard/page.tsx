import Link from "next/link";

import { requireUser } from "@/features/auth/server/require-user";
import { DashboardRefreshButton } from "@/features/listings/components/dashboard-refresh-button";
import { getSellerListingsPage } from "@/features/listings/server/get-seller-listings";

export default async function DashboardOverviewPage() {
  await requireUser();
  const listingsPage = await getSellerListingsPage();

  return (
    <div className="flex flex-col gap-8" data-testid="dashboard-shell">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-slate-900">Dashboard</h1>
        <p className="mt-2 text-slate-500">Welcome to your seller dashboard. Manage your listings and inventory.</p>
      </div>

      <section className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {/* Card 1 */}
        <div className="rounded-xl border border-slate-200 bg-white text-slate-950 shadow-sm">
          <div className="flex flex-row items-center justify-between p-6 pb-2">
            <h3 className="text-sm font-medium tracking-tight text-slate-500">Total Listings</h3>
          </div>
          <div className="p-6 pt-0">
            <div className="text-2xl font-bold">{listingsPage.total}</div>
          </div>
        </div>

        {/* Card 2 */}
        <div className="rounded-xl border border-slate-200 bg-white text-slate-950 shadow-sm">
          <div className="flex flex-row items-center justify-between p-6 pb-2">
            <h3 className="text-sm font-medium tracking-tight text-slate-500">Active Page</h3>
          </div>
          <div className="p-6 pt-0">
            <div className="text-2xl font-bold">{listingsPage.page}</div>
          </div>
        </div>

        {/* Card 3 */}
        <div className="rounded-xl border border-slate-200 bg-white text-slate-950 shadow-sm md:col-span-2 lg:col-span-1">
          <div className="flex flex-row items-center justify-between p-6 pb-2">
            <h3 className="text-sm font-medium tracking-tight text-slate-500">Workflow</h3>
          </div>
          <div className="p-6 pt-0">
            <p className="text-sm text-slate-500">Create, update, and curate listing images while the backend stays the auth and data authority.</p>
          </div>
        </div>
      </section>

      <section className="flex flex-wrap items-center gap-4">
        <DashboardRefreshButton />
        <Link className="inline-flex h-9 items-center justify-center rounded-md border border-slate-200 bg-white px-4 text-sm font-medium text-slate-900 shadow-sm transition-colors hover:bg-slate-100 hover:text-slate-900" href="/dashboard/listings">
          Review listings table
        </Link>
        <Link className="inline-flex h-9 items-center justify-center rounded-md bg-slate-900 px-4 text-sm font-medium text-slate-50 shadow transition-colors hover:bg-slate-900/90" href="/dashboard/listings/new">
          Create listing
        </Link>
      </section>
    </div>
  );
}
