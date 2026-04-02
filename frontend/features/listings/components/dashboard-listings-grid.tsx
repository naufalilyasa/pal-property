"use client";

import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";

import type { SellerListing } from "@/lib/api/seller-listings";
import { deleteSellerListing, updateSellerListingStatus } from "@/lib/api/seller-listings-client";

function formatPrice(price: number, currency: string) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: currency || "IDR",
    maximumFractionDigits: 0,
  }).format(price);
}

function formatStatus(status: string) {
  const labels: Record<string, string> = {
    active: "Aktif",
    inactive: "Tidak aktif",
    sold: "Terjual",
    draft: "Draft",
    archived: "Arsip",
    published: "Aktif",
  };

  return labels[status] ?? status;
}

function getStatusClassName(status: string) {
  if (status === "sold") return "bg-emerald-100 text-emerald-800";
  if (status === "archived") return "bg-slate-200 text-slate-700";
  if (status === "draft") return "bg-amber-100 text-amber-800";
  return "bg-blue-100 text-blue-800";
}

export function DashboardListingsGrid({ listings: initialListings }: { listings: SellerListing[] }) {
  const router = useRouter();
  const [listings, setListings] = useState(initialListings);
  const [busyAction, setBusyAction] = useState<string | null>(null);

  const sortedListings = useMemo(
    () => [...listings].sort((left, right) => new Date(right.updated_at).getTime() - new Date(left.updated_at).getTime()),
    [listings],
  );

  const runStatusAction = async (listing: SellerListing, status: "sold" | "archived") => {
    setBusyAction(`${listing.id}:${status}`);
    try {
      const updated = await updateSellerListingStatus(listing.id, status);
      setListings((current) => current.map((candidate) => (candidate.id === listing.id ? updated : candidate)));
      router.refresh();
    } finally {
      setBusyAction(null);
    }
  };

  const runDeleteAction = async (listing: SellerListing) => {
    if (!window.confirm(`Delete listing "${listing.title}"? This action cannot be undone.`)) {
      return;
    }

    setBusyAction(`${listing.id}:delete`);
    try {
      await deleteSellerListing(listing.id);
      setListings((current) => current.filter((candidate) => candidate.id !== listing.id));
      router.refresh();
    } finally {
      setBusyAction(null);
    }
  };

  return (
    <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3" data-testid="dashboard-listings-grid">
      {sortedListings.map((listing) => {
        const image = listing.images.find((candidate) => candidate.is_primary) ?? listing.images[0] ?? null;
        const location = [listing.location_district, listing.location_city].filter(Boolean).join(", ");
        const isSold = listing.status === "sold";
        const isArchived = listing.status === "archived";

        return (
          <article className="overflow-hidden rounded-[1.75rem] border border-slate-200 bg-white/90 shadow-sm" data-testid={`dashboard-listing-card-${listing.id}`} key={listing.id}>
            <div className="relative h-56 overflow-hidden bg-slate-50">
              {image ? (
                <Image alt={listing.title} fill sizes="(min-width: 1280px) 25vw, (min-width: 768px) 50vw, 100vw" src={image.url} className="object-cover" unoptimized />
              ) : (
                <div className="flex h-full items-center justify-center px-3 text-center text-sm text-slate-900">No image</div>
              )}
            </div>

            <div className="space-y-4 p-5">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <h2 className="line-clamp-2 text-lg font-semibold text-slate-900">{listing.title}</h2>
                  <p className="mt-1 text-sm text-slate-500">{location || "Lokasi belum diisi"}</p>
                </div>
                <span className={`rounded-full px-3 py-1 text-[11px] font-semibold uppercase tracking-wider ${getStatusClassName(listing.status)}`}>
                  {formatStatus(listing.status)}
                </span>
              </div>

              <div className="flex items-center justify-between gap-3">
                <p className="text-base font-semibold text-slate-900">{formatPrice(listing.price, listing.currency)}</p>
                <span className="rounded-full bg-slate-100 px-3 py-1 text-[11px] font-semibold uppercase tracking-wider text-slate-700">
                  {listing.category?.name ?? "Uncategorized"}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-2">
                <Link className="inline-flex items-center justify-center rounded-full border border-slate-200 bg-slate-50 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:border-slate-900" href={`/dashboard/listings/${listing.id}/edit`}>
                  Edit listing
                </Link>
                <button
                  className="inline-flex items-center justify-center rounded-full border border-emerald-200 bg-emerald-50 px-4 py-2 text-sm font-semibold text-emerald-700 transition hover:border-emerald-500 disabled:cursor-not-allowed disabled:opacity-50"
                  disabled={isSold || busyAction !== null}
                  onClick={() => runStatusAction(listing, "sold")}
                  type="button"
                >
                  {busyAction === `${listing.id}:sold` ? "Saving..." : "Mark sold"}
                </button>
                <button
                  className="inline-flex items-center justify-center rounded-full border border-slate-200 bg-slate-50 px-4 py-2 text-sm font-semibold text-slate-700 transition hover:border-slate-500 disabled:cursor-not-allowed disabled:opacity-50"
                  disabled={isArchived || busyAction !== null}
                  onClick={() => runStatusAction(listing, "archived")}
                  type="button"
                >
                  {busyAction === `${listing.id}:archived` ? "Saving..." : "Archive"}
                </button>
                <button
                  className="inline-flex items-center justify-center rounded-full border border-red-200 bg-red-50 px-4 py-2 text-sm font-semibold text-red-700 transition hover:border-red-500 disabled:cursor-not-allowed disabled:opacity-50"
                  disabled={busyAction !== null}
                  onClick={() => runDeleteAction(listing)}
                  type="button"
                >
                  {busyAction === `${listing.id}:delete` ? "Deleting..." : "Delete"}
                </button>
              </div>
            </div>
          </article>
        );
      })}
    </div>
  );
}
