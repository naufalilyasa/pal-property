import Image from "next/image";
import Link from "next/link";

import type { ListingRecord } from "@/lib/api/listing-form";

function formatPrice(price: number, currency: string) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: currency || "IDR",
    maximumFractionDigits: 0,
  }).format(price);
}

export function ListingCard({ href, listing }: { href: string; listing: ListingRecord }) {
  const image = listing.images[0] ?? null;

  return (
    <article className="overflow-hidden rounded-[1.75rem] border border-[var(--line)] bg-white/80">
      <div className="relative h-52 bg-[var(--panel)]">
        {image ? (
          <Image alt={listing.title} fill sizes="(min-width: 1024px) 33vw, 100vw" src={image.url} className="object-cover" unoptimized />
        ) : (
          <div className="flex h-full items-center justify-center px-6 text-center text-sm text-[var(--muted)]">No image uploaded yet</div>
        )}
      </div>
      <div className="space-y-4 p-5">
        <div className="space-y-2">
          <p className="text-xs uppercase tracking-[0.24em] text-[var(--accent)]">{listing.category?.name ?? "Uncategorized"}</p>
          <h2 className="text-2xl font-semibold tracking-[-0.03em] text-[var(--ink)]">{listing.title}</h2>
          <p className="text-sm leading-7 text-[var(--muted)]">{listing.location_city ?? "Unknown city"} · {listing.status}</p>
        </div>
        <div className="flex items-center justify-between gap-4">
          <p className="text-lg font-semibold text-[var(--ink)]">{formatPrice(listing.price, listing.currency)}</p>
          <Link className="rounded-full border border-[var(--line)] bg-[var(--panel)] px-4 py-2 text-sm font-semibold text-[var(--ink)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]" href={href}>
            View details
          </Link>
        </div>
      </div>
    </article>
  );
}
