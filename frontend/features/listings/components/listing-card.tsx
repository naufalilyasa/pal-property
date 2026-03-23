import Image from "next/image";
import Link from "next/link";

import type { ListingSpecifications } from "@/lib/api/listing-form";
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
  const specifications = parseSpecifications(listing.specifications);
  const locationLabel = [listing.location_district, listing.location_city].filter(Boolean).join(", ") || "Indonesia";
  const statItems = [
    specifications.bedrooms ? `${specifications.bedrooms} bd` : null,
    specifications.bathrooms ? `${specifications.bathrooms} ba` : null,
    specifications.building_area_sqm ? `${specifications.building_area_sqm} m2` : null,
  ].filter(Boolean);

  return (
    <article className="group space-y-2.5 bg-transparent">
      <div className="relative h-44 overflow-hidden rounded-sm bg-[#ece9e2] md:h-40 xl:h-44">
        {image ? (
          <Image alt={listing.title} fill sizes="(min-width: 1280px) 22vw, (min-width: 768px) 30vw, 100vw" src={image.url} className="object-cover transition duration-300 group-hover:scale-[1.03]" unoptimized />
        ) : (
          <div className="flex h-full items-center justify-center px-6 text-center text-sm text-[var(--muted)]">No image uploaded yet</div>
        )}
        <div className="absolute left-2 top-2 rounded-full bg-[rgba(255,255,255,0.94)] px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.18em] text-[var(--ink)]">
          {listing.is_featured ? "Featured" : listing.category?.name ?? "Property"}
        </div>
      </div>
      <div className="space-y-1.5 px-0.5">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-[1.05rem] font-semibold tracking-[-0.03em] text-[#111]">{formatPrice(listing.price, listing.currency)}</p>
            <p className="mt-0.5 text-[10px] font-medium uppercase tracking-[0.16em] text-[#76736d]">
              {(listing.category?.name ?? "House") + " • " + listing.status}
            </p>
          </div>
          <p className="text-[10px] font-medium uppercase tracking-[0.16em] text-[#76736d]">{listing.view_count} views</p>
        </div>
        <div className="flex flex-wrap gap-1 text-[11px] text-[#2b2b29]">
          {(statItems.length > 0 ? statItems : ["Move-in ready"]).map((item) => (
            <span key={item}>{item}</span>
          ))}
        </div>
        <p className="line-clamp-1 text-[11px] text-[#6d6a64]">{locationLabel}</p>
        <p className="line-clamp-1 text-[11px] text-[#6d6a64]">{listing.title}</p>
        <div className="flex items-center gap-2 pt-1.5">
          <button aria-label="Save listing" className="flex h-7 w-7 items-center justify-center rounded-full border border-[#d8d5cf] text-[11px] text-[#444]" type="button">
            +
          </button>
          <button aria-label="Share listing" className="flex h-7 w-7 items-center justify-center rounded-full border border-[#d8d5cf] text-[11px] text-[#444]" type="button">
            o
          </button>
          <Link className="ml-auto text-[11px] font-semibold uppercase tracking-[0.16em] text-[var(--ink)] hover:text-[var(--accent)]" href={href}>
            Details
          </Link>
        </div>
      </div>
    </article>
  );
}

function parseSpecifications(specifications: unknown): Partial<ListingSpecifications> {
  if (!specifications || typeof specifications !== "object") {
    return {};
  }

  const candidate = specifications as Record<string, unknown>;

  return {
    bedrooms: toNumber(candidate.bedrooms),
    bathrooms: toNumber(candidate.bathrooms),
    land_area_sqm: toNumber(candidate.land_area_sqm),
    building_area_sqm: toNumber(candidate.building_area_sqm),
  };
}

function toNumber(value: unknown): number | undefined {
  return typeof value === "number" && Number.isFinite(value) ? value : undefined;
}
