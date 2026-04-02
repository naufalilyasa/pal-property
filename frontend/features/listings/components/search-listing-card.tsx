"use client";

import Image from "next/image";
import Link from "next/link";
import { useMemo, useState } from "react";

import { SaveListingButton } from "@/features/saved-listings/components/save-listing-button";
import type { SearchListingCard } from "@/features/listings/server/get-search-listings";

function formatPrice(price: number, currency: string) {
  const normalizedCurrency = currency || "IDR";

  if (normalizedCurrency === "IDR") {
    if (price >= 1_000_000_000) {
      return `Rp ${formatCompactNumber(price / 1_000_000_000)} Miliar`;
    }

    if (price >= 1_000_000) {
      return `Rp ${formatCompactNumber(price / 1_000_000)} Juta`;
    }
  }

  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: normalizedCurrency,
    maximumFractionDigits: 0,
  }).format(price);
}

function formatCompactNumber(value: number) {
  return new Intl.NumberFormat("id-ID", {
    minimumFractionDigits: value < 10 && !Number.isInteger(value) ? 1 : 0,
    maximumFractionDigits: value < 10 ? 1 : 0,
  }).format(value);
}

function formatMetric(label: string, value?: number) {
  return typeof value === "number" ? `${label} ${value}` : null;
}

function formatArea(label: string, value?: number) {
  return typeof value === "number" ? `${label} ${value} m²` : null;
}

function formatTransactionType(value: string) {
  return value === "rent" ? "Sewa" : value === "sale" ? "Jual" : value;
}

export function SearchListingCardItem({
  href,
  listing,
  initialSaved = false,
  refreshOnRemove = false,
}: {
  href: string;
  listing: SearchListingCard;
  initialSaved?: boolean;
  refreshOnRemove?: boolean;
}) {
  const images = useMemo(() => {
    const ordered = [listing.primary_image_url, ...(listing.image_urls ?? [])].filter(
      (value): value is string => typeof value === "string" && value.length > 0,
    );

    return [...new Set(ordered)];
  }, [listing.image_urls, listing.primary_image_url]);
  const visibleImages = images.slice(0, 5);
  const [activeIndex, setActiveIndex] = useState(0);
  const image = visibleImages[activeIndex] ?? null;
  const additionalImages = Math.max(images.length - visibleImages.length, 0);
  const price = formatPrice(listing.price, listing.currency);
  const primaryLocation = [listing.location_village, listing.location_district, listing.location_city].filter(Boolean).join(", ");
  const secondaryLocation = [listing.location_province].filter(Boolean).join(", ");
  const badges = [
    listing.category?.name,
    formatTransactionType(listing.transaction_type),
    listing.status,
    listing.is_featured ? "Featured" : null,
  ].filter(Boolean);
  const metrics = [
    formatMetric("KT", listing.bedroom_count),
    formatMetric("KM", listing.bathroom_count),
    formatArea("LT", listing.land_area_sqm),
    formatArea("LB", listing.building_area_sqm),
  ].filter(Boolean);

  const canSlide = visibleImages.length > 1;
  const showMoreOverlay = additionalImages > 0 && activeIndex === visibleImages.length - 1;

  return (
    <article data-testid="listing-card" className="group flex flex-col space-y-3 bg-white">
      <div className="relative">
        <Link href={href} className="relative block aspect-4/3 w-full overflow-hidden rounded-xl bg-gray-100">
          {image ? (
            <Image
              alt={listing.title}
              fill
              sizes="(min-width: 1280px) 25vw, (min-width: 768px) 33vw, 100vw"
              src={image}
              className="object-contain bg-gray-50 p-2 transition duration-300 group-hover:scale-[1.01]"
              unoptimized
            />
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-gray-500">No image</div>
          )}

          {showMoreOverlay ? (
            <div className="absolute inset-0 flex flex-col items-center justify-center bg-black/45 px-6 text-center text-white">
              <span className="text-2xl font-semibold tracking-tight">+{additionalImages}</span>
              <span className="mt-1 text-sm font-medium">more photos</span>
              <span className="mt-2 text-xs uppercase tracking-[0.2em] text-white/80">Open detail listing</span>
            </div>
          ) : null}

          {canSlide ? (
            <div className="absolute bottom-3 left-1/2 z-10 flex -translate-x-1/2 items-center gap-1 rounded-full bg-black/45 px-2 py-1">
              {visibleImages.map((visibleImage, index) => (
                <span
                  className={index === activeIndex ? "h-2.5 w-2.5 rounded-full bg-white" : "h-2 w-2 rounded-full bg-white/45"}
                  key={`${listing.id}-dot-${visibleImage}`}
                />
              ))}
            </div>
          ) : null}
        </Link>

        {canSlide ? (
          <>
            <button
              aria-label={`Show previous image for ${listing.title}`}
              className="absolute left-3 top-1/2 z-10 flex h-9 w-9 -translate-y-1/2 items-center justify-center rounded-full bg-white/90 text-sm font-semibold text-[#111] shadow-sm transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-40"
              disabled={activeIndex === 0}
              onClick={() => setActiveIndex((current) => Math.max(0, current - 1))}
              type="button"
            >
              ‹
            </button>
            <button
              aria-label={`Show next image for ${listing.title}`}
              className="absolute right-3 top-1/2 z-10 flex h-9 w-9 -translate-y-1/2 items-center justify-center rounded-full bg-white/90 text-sm font-semibold text-[#111] shadow-sm transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-40"
              disabled={activeIndex === visibleImages.length - 1}
              onClick={() => setActiveIndex((current) => Math.min(visibleImages.length - 1, current + 1))}
              type="button"
            >
              ›
            </button>
          </>
        ) : null}

        <div className="absolute right-3 top-3 z-10">
          <SaveListingButton
            initialSaved={initialSaved}
            listingId={listing.id}
            refreshOnRemove={refreshOnRemove}
            scope="repeated"
            variant="icon"
          />
        </div>
      </div>

      <div className="flex flex-col space-y-2">
        <Link href={href} className="text-lg font-bold tracking-tight text-[#111] hover:underline">
          {price}
        </Link>
        {badges.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {badges.map((badge) => (
              <span
                key={`${listing.id}-${badge}`}
                className="rounded-full bg-[#eef1fb] px-2.5 py-1 text-[11px] font-semibold text-[#3a4aa1]"
              >
                {badge}
              </span>
            ))}
          </div>
        ) : null}
        <p className="text-sm font-semibold text-[#111]">{listing.title}</p>
        {primaryLocation ? <p className="line-clamp-1 text-[12px] font-medium text-[#4b5563]">{primaryLocation}</p> : null}
        {secondaryLocation ? <p className="line-clamp-1 text-[12px] text-[#6d6a64]">{secondaryLocation}</p> : null}
        {listing.description_excerpt ? <p className="line-clamp-2 text-[12px] text-[#6d6a64]">{listing.description_excerpt}</p> : null}
        {metrics.length > 0 ? (
          <div className="flex flex-wrap gap-2 pt-1">
            {metrics.map((metric) => (
              <span
                key={`${listing.id}-${metric}`}
                className="rounded-full border border-[#e6e7eb] bg-[#f8f8fa] px-2.5 py-1 text-[11px] font-medium text-[#374151]"
              >
                {metric}
              </span>
            ))}
          </div>
        ) : null}
      </div>
    </article>
  );
}
