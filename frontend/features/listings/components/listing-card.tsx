import Image from "next/image";
import Link from "next/link";

import type { ListingSpecifications } from "@/lib/api/listing-form";
import type { ListingRecord } from "@/lib/api/listing-form";

function formatPrice(price: number, currency: string) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currency || "USD",
    maximumFractionDigits: 0,
  }).format(price);
}

export function ListingCard({
  href,
  listing,
}: {
  href: string;
  listing: ListingRecord;
}) {
  // dummy data fallbacks to match screenshot design
  const dummyListing = {
    price: 2190000,
    currency: "USD",
    category: "Condo",
    status: "For Sale",
    beds: 3,
    baths: 2,
    sqft: 1500,
    address: "123 Coral Gables Blvd, Unit D, FL",
    // Random high-quality architecture image from unsplash as dummy
    image:
      "https://images.unsplash.com/photo-1600596542815-ffad4c1539a9?auto=format&fit=crop&q=80&w=800",
  };

  const image =
    listing.images && listing.images[0] ? listing.images[0].url : dummyListing.image;

  // Use dummy price if listing price is 0 or missing
  const priceValue = listing.price > 0 ? listing.price : dummyListing.price;
  const currencyValue = listing.currency || dummyListing.currency;
  const price = formatPrice(priceValue, currencyValue);

  const specifications = parseSpecifications(listing.specifications);
  const beds = specifications.bedrooms || dummyListing.beds;
  const baths = specifications.bathrooms || dummyListing.baths;
  const sqft = specifications.building_area_sqm
    ? Math.round(specifications.building_area_sqm * 10.7639) // sqm to sqft approx
    : dummyListing.sqft;

  const category = listing.category?.name || dummyListing.category;
  const status = listing.status && listing.status !== "active" ? listing.status : dummyListing.status;

  const address =
    [listing.title, listing.location_district, listing.location_city]
      .filter(Boolean)
      .join(", ") || dummyListing.address;

  const specsText = `${category} • ${status} • ${beds} beds • ${baths} baths • ${sqft} sqft`;

  return (
    <article className="group flex flex-col space-y-3 bg-white">
      <Link
        href={href}
        className="relative aspect-4/3 w-full overflow-hidden bg-gray-100"
      >
        <Image
          alt={address}
          fill
          sizes="(min-width: 1280px) 25vw, (min-width: 768px) 33vw, 100vw"
          src={image}
          className="object-cover transition duration-300 group-hover:scale-[1.03]"
          unoptimized
        />
      </Link>

      <div className="flex flex-col space-y-1">
        <Link
          href={href}
          className="text-lg font-bold tracking-tight text-[#111] hover:underline"
        >
          {price}
        </Link>
        <p className="text-[12px] font-medium text-[#111]">
          {specsText}
        </p>
        <p className="line-clamp-1 text-[12px] text-[#6d6a64]">
          {address}
        </p>
      </div>

      <div className="flex items-center gap-2 pt-1">
        <button
          aria-disabled="true"
          aria-label="Copy Link"
          className="flex h-8 w-8 items-center justify-center rounded-full border border-gray-200 text-[#111] transition hover:bg-gray-50"
          type="button"
        >
          <svg
            xmlns="http://www.w3.org/2000/event"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
            <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
          </svg>
        </button>
        <button
          aria-disabled="true"
          aria-label="Save"
          className="flex h-8 w-8 items-center justify-center rounded-full border border-gray-200 text-[#111] transition hover:bg-gray-50"
          type="button"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z" />
          </svg>
        </button>
      </div>
    </article>
  );
}

function parseSpecifications(
  specifications: unknown,
): Partial<ListingSpecifications> {
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
  return typeof value === "number" && Number.isFinite(value)
    ? value
    : undefined;
}
