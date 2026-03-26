import Image from "next/image";
import Link from "next/link";

import type { SearchListingCard } from "@/features/listings/server/get-search-listings";

function formatPrice(price: number, currency: string) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currency || "IDR",
    maximumFractionDigits: 0,
  }).format(price);
}

export function SearchListingCardItem({ href, listing }: { href: string; listing: SearchListingCard }) {
  const image = listing.primary_image_url ?? listing.image_urls?.[0] ?? null;
  const price = formatPrice(listing.price, listing.currency);
  const address = [listing.location_district, listing.location_city, listing.location_province].filter(Boolean).join(", ");
  const meta = [listing.category?.name, listing.transaction_type, listing.status].filter(Boolean).join(" • ");

  return (
    <article data-testid="listing-card" className="group flex flex-col space-y-3 bg-white">
      <Link href={href} className="relative aspect-4/3 w-full overflow-hidden rounded-xl bg-gray-100">
        {image ? (
          <Image
            alt={listing.title}
            fill
            sizes="(min-width: 1280px) 25vw, (min-width: 768px) 33vw, 100vw"
            src={image}
            className="object-cover transition duration-300 group-hover:scale-[1.03]"
            unoptimized
          />
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-gray-500">No image</div>
        )}
      </Link>

      <div className="flex flex-col space-y-1">
        <Link href={href} className="text-lg font-bold tracking-tight text-[#111] hover:underline">
          {price}
        </Link>
        <p className="text-[12px] font-medium text-[#111]">{meta}</p>
        <p className="text-sm font-semibold text-[#111]">{listing.title}</p>
        {listing.description_excerpt ? <p className="line-clamp-2 text-[12px] text-[#6d6a64]">{listing.description_excerpt}</p> : null}
        {address ? <p className="line-clamp-1 text-[12px] text-[#6d6a64]">{address}</p> : null}
      </div>
    </article>
  );
}
