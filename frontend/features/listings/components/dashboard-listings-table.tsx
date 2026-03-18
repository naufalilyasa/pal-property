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
    <div className="overflow-hidden rounded-[1.75rem] border border-[var(--line)] bg-white/80" data-testid="dashboard-listings-table">
      <div className="grid gap-0">
        {listings.map((listing) => {
          const image = listing.images.find((candidate) => candidate.is_primary) ?? listing.images[0] ?? null;

          return (
            <article className="grid gap-4 border-b border-[var(--line)] p-5 last:border-b-0 lg:grid-cols-[120px_1fr_auto] lg:items-center" key={listing.id}>
              <div className="relative h-24 overflow-hidden rounded-[1rem] bg-[var(--panel)]">
                {image ? (
                  <Image alt={listing.title} fill sizes="120px" src={image.url} className="object-cover" unoptimized />
                ) : (
                  <div className="flex h-full items-center justify-center px-3 text-center text-xs text-[var(--muted)]">No image</div>
                )}
              </div>
              <div className="space-y-2">
                <h2 className="text-lg font-semibold text-[var(--ink)]">{listing.title}</h2>
                <p className="text-sm text-[var(--muted)]">{listing.category?.name ?? "Uncategorized"} · {listing.status} · {listing.location_city ?? "Unknown city"}</p>
              </div>
              <div className="flex flex-col items-start gap-3 lg:items-end">
                <p className="text-sm font-semibold text-[var(--ink)]">{formatPrice(listing.price, listing.currency)}</p>
                <Link className="rounded-full border border-[var(--line)] bg-[var(--panel)] px-4 py-2 text-sm font-semibold text-[var(--ink)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]" href={`/dashboard/listings/${listing.id}/edit`}>
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
