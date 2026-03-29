import Link from "next/link";

import { requireUser } from "@/features/auth/server/require-user";
import { Footer } from "@/features/listings/components/footer";
import { SearchListingCardItem } from "@/features/listings/components/search-listing-card";
import { TopNav } from "@/features/listings/components/top-nav";
import type { SearchListingCard } from "@/features/listings/server/get-search-listings";
import { getSavedListingsPage } from "@/features/saved-listings/server/get-saved-listings";
import type { SavedListing } from "@/lib/api/saved-listings";

const DEFAULT_PAGE = "1";
const DEFAULT_LIMIT = "12";

export default async function SavedListingsPage({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
}) {
  const [resolvedSearchParams] = await Promise.all([searchParams, requireUser()]);
  const normalizedSearchParams = resolvedSearchParams ?? {};
  const page = getQueryValue(normalizedSearchParams.page) ?? DEFAULT_PAGE;
  const limit = getQueryValue(normalizedSearchParams.limit) ?? DEFAULT_LIMIT;

  const savedListingsPage = await getSavedListingsPage({ page, limit });
  const listings = savedListingsPage.data.map(mapSavedListingToCard);

  return (
    <div className="flex min-h-screen flex-col bg-stone-50 text-[#111]">
      <TopNav />

      <main className="flex-1">
        <section className="border-b border-stone-200 bg-[radial-gradient(circle_at_top_left,_rgba(24,24,27,0.08),_transparent_42%),linear-gradient(180deg,_rgba(255,255,255,0.95),_rgba(250,250,249,0.92))]">
          <div className="mx-auto flex max-w-[1580px] flex-col gap-6 px-4 py-12 sm:px-6 lg:px-8 lg:py-16">
            <div className="flex flex-col gap-3">
              <p className="text-[11px] font-semibold uppercase tracking-[0.3em] text-stone-500">
                Buyer workspace
              </p>
              <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
                <div className="max-w-2xl space-y-3">
                  <h1 className="text-3xl font-bold tracking-tight text-stone-950 sm:text-4xl">
                    Saved listings, ready when you are.
                  </h1>
                  <p className="max-w-xl text-sm leading-6 text-stone-600 sm:text-base">
                    Your shortlist stays protected and server-rendered, with the newest saves surfaced first so you can
                    jump back into active properties quickly.
                  </p>
                </div>

                <div className="flex flex-wrap items-center gap-3 text-sm text-stone-600">
                  <span className="rounded-full border border-stone-200 bg-white px-4 py-2 font-semibold text-stone-900">
                    {savedListingsPage.total.toLocaleString("en-US")} saved
                  </span>
                  <span className="rounded-full border border-stone-200 px-4 py-2">Newest saved first</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="mx-auto w-full max-w-[1580px] px-4 py-10 sm:px-6 lg:px-8 lg:py-12">
          {listings.length === 0 ? (
            <div
              className="rounded-[28px] border border-dashed border-stone-300 bg-white px-6 py-16 text-center shadow-sm sm:px-10"
              data-testid="saved-listings-empty"
            >
              <p className="text-[11px] font-semibold uppercase tracking-[0.3em] text-stone-500">No saved listings yet</p>
              <h2 className="mt-4 text-2xl font-bold tracking-tight text-stone-950">Start building your shortlist from live listings.</h2>
              <p className="mx-auto mt-3 max-w-xl text-sm leading-6 text-stone-600">
                Browse the latest homes, tap save on anything worth revisiting, and this page will keep the collection organized for you.
              </p>
              <Link
                className="mt-8 inline-flex h-11 items-center justify-center rounded-full bg-stone-950 px-6 text-sm font-semibold text-white transition hover:bg-stone-800"
                href="/listings"
              >
                Explore listings
              </Link>
            </div>
          ) : (
            <div className="space-y-10">
              <div className="grid gap-x-4 gap-y-8 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4" data-testid="saved-listings-grid">
                {listings.map((listing) => (
                  <SearchListingCardItem
                    key={listing.id}
                    href={`/listings/${listing.slug}`}
                    initialSaved
                    listing={listing}
                    refreshOnRemove
                  />
                ))}
              </div>

              {savedListingsPage.total_pages > 1 ? (
                <div className="flex items-center justify-center gap-2" data-testid="saved-listings-pagination">
                  {Array.from({ length: savedListingsPage.total_pages }, (_, index) => index + 1).map((pageNumber) => {
                    const params = new URLSearchParams();
                    params.set("page", String(pageNumber));
                    params.set("limit", limit);

                    const isCurrent = pageNumber === savedListingsPage.page;

                    return (
                      <Link
                        key={pageNumber}
                        className={`flex h-8 min-w-8 items-center justify-center rounded-full px-3 text-sm font-semibold transition ${
                          isCurrent
                            ? "bg-black text-white"
                            : "border border-gray-300 text-[#111] hover:bg-gray-50"
                        }`}
                        href={`/saved-listings?${params.toString()}`}
                      >
                        {pageNumber}
                      </Link>
                    );
                  })}
                </div>
              ) : null}
            </div>
          )}
        </section>
      </main>

      <Footer />
    </div>
  );
}

function mapSavedListingToCard(listing: SavedListing): SearchListingCard {
  const primaryImage = listing.images.find((image) => image.is_primary)?.url ?? listing.images[0]?.url;

  return {
    id: listing.id,
    category_id: listing.category_id ?? undefined,
    category: listing.category
      ? {
          id: listing.category.id,
          name: listing.category.name,
          slug: listing.category.slug,
        }
      : undefined,
    title: listing.title,
    slug: listing.slug,
    description_excerpt: listing.description ?? undefined,
    transaction_type: listing.transaction_type,
    price: listing.price,
    currency: listing.currency,
    location_city: listing.location_city ?? undefined,
    location_district: listing.location_district ?? undefined,
    status: listing.status,
    is_featured: listing.is_featured,
    primary_image_url: primaryImage,
    image_urls: listing.images.map((image) => image.url),
    created_at: listing.created_at,
    updated_at: listing.updated_at,
  };
}

function getQueryValue(value: string | string[] | undefined) {
  if (Array.isArray(value)) {
    return value[0];
  }

  return value;
}
