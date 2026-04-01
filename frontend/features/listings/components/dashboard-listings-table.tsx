import Image from "next/image";
import Link from "next/link";

import type { SellerListing } from "@/lib/api/seller-listings";

function formatPrice(price: number, currency: string) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: currency || "IDR",
    maximumFractionDigits: 0,
  }).format(price);
}

export function DashboardListingsTable({ listings }: { listings: SellerListing[] }) {
  return (
    <div className="overflow-hidden rounded-[1.75rem] border border-slate-200 bg-white/80" data-testid="dashboard-listings-table">
      <div className="grid gap-0">
        {listings.map((listing) => {
          const image = listing.images.find((candidate) => candidate.is_primary) ?? listing.images[0] ?? null;

          return (
            <article className="grid gap-4 border-b border-slate-200 p-5 last:border-b-0 lg:grid-cols-[120px_1fr_auto] lg:items-center" key={listing.id}>
              <div className="relative h-24 overflow-hidden rounded-[1rem] bg-slate-50">
                {image ? (
                  <Image alt={listing.title} fill sizes="120px" src={image.url} className="object-cover" unoptimized />
                ) : (
                  <div className="flex h-full items-center justify-center px-3 text-center text-xs text-slate-900">No image</div>
                )}
              </div>
              <div className="space-y-2">
                <h2 className="text-lg font-semibold text-slate-900">{listing.title}</h2>
                <p className="text-sm text-slate-900">{listing.category?.name ?? "Uncategorized"} · {listing.status} · {listing.location_city ?? "Unknown city"}</p>
              </div>
              <div className="flex flex-col items-start gap-3 lg:items-end">
                <p className="text-sm font-semibold text-slate-900">{formatPrice(listing.price, listing.currency)}</p>
                <Link className="rounded-full border border-slate-200 bg-slate-50 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:border-slate-900 hover:text-slate-900" href={`/dashboard/listings/${listing.id}/edit`}>
                  Edit listing
                </Link>
              </div>
            </article>
          );
        })}
      </div>
    </div>
  );
}
