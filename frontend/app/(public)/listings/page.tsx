import Link from "next/link";

import { ListingFilters } from "@/features/listings/components/listing-filters";
import { SearchListingCardItem } from "@/features/listings/components/search-listing-card";
import { getSearchListings } from "@/features/listings/server/get-search-listings";
import { TopNav } from "@/features/listings/components/top-nav";
import { Footer } from "@/features/listings/components/footer";

export default async function PublicListingsPage({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
}) {
  const resolvedSearchParams = (await searchParams) ?? {};

  // View mode
  const view =
    getQueryValue(resolvedSearchParams.view) === "list" ? "list" : "map";

  // Filter values
  const filterValues = {
    q: getQueryValue(resolvedSearchParams.q),
    transaction_type: getQueryValue(resolvedSearchParams.transaction_type),
    category_id: getQueryValue(resolvedSearchParams.category_id),
    location_province: getQueryValue(resolvedSearchParams.location_province),
    location_city: getQueryValue(resolvedSearchParams.location_city),
    price_min: getQueryValue(resolvedSearchParams.price_min),
    price_max: getQueryValue(resolvedSearchParams.price_max),
    sort: getQueryValue(resolvedSearchParams.sort),
    limit: getQueryValue(resolvedSearchParams.limit) ?? "12",
  };

  const listingsPage = await getSearchListings({
    page: getQueryValue(resolvedSearchParams.page) ?? "1",
    limit: filterValues.limit,
    q: filterValues.q,
    transaction_type: filterValues.transaction_type,
    category_id: filterValues.category_id,
    location_province: filterValues.location_province,
    location_city: filterValues.location_city,
    price_min: filterValues.price_min,
    price_max: filterValues.price_max,
    sort: filterValues.sort,
  });
  const displayListings = listingsPage.items;
  const displayTotal = listingsPage.total;

  return (
    <div className="flex min-h-screen flex-col bg-white font-sans text-[#111]">
      <TopNav />
      {/* Sticky Filter Bar */}
      <div className="sticky top-0 z-10">
        <ListingFilters view={view} />
      </div>

      {/* Main Content */}
      <main className="flex flex-1">
        {view === "map" && (
          <div data-testid="listing-map-panel" className="hidden w-[45%] border-r border-gray-200 lg:block">
            <div className="sticky top-[68px] h-[calc(100vh-68px)] w-full">
              <iframe
                className="h-full w-full border-0 grayscale-20 opacity-90"
                loading="lazy"
                referrerPolicy="no-referrer-when-downgrade"
                src="https://www.openstreetmap.org/export/embed.html?bbox=106.5608%2C-6.3754%2C107.0217%2C-5.9949&layer=mapnik"
                title="Property search map"
              />
            </div>
          </div>
        )}

        <div
          className={`p-4 pb-20 sm:p-6 ${view === "map" ? "w-full lg:w-[55%]" : "mx-auto w-full max-w-[1580px]"
            }`}
        >
          {/* Listings Header */}
          <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 className="text-sm font-semibold">
                {displayTotal.toLocaleString("en-US")} Results
              </h2>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <span className="text-gray-500">Sort:</span>
              <select className="bg-transparent font-semibold text-[#111] outline-none">
                <option>Newest</option>
                <option>Price: Low to High</option>
                <option>Price: High to Low</option>
              </select>
            </div>
          </div>

          {/* Listings Grid */}
          <div
            className={`grid gap-x-4 gap-y-8 ${view === "map"
                ? "grid-cols-1 sm:grid-cols-2"
                : "grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
              }`}
          >
            {displayListings.length === 0 ? (
              <div className="col-span-full rounded-md border border-dashed border-gray-300 bg-gray-50 p-12 text-center">
                <p className="text-sm font-semibold uppercase tracking-widest text-gray-500">
                  No results
                </p>
                <h3 className="mt-3 text-xl font-bold tracking-tight text-[#111]">
                  No properties matched your search
                </h3>
                <p className="mt-2 text-sm text-gray-600">
                  Adjust filters, query text, or sort to explore more listings.
                </p>
              </div>
            ) : null}

            {displayListings.map((listing) => (
              <SearchListingCardItem
                key={listing.id}
                href={`/listings/${listing.slug}`}
                listing={listing}
              />
            ))}
          </div>

          {/* Pagination */}
          {listingsPage.total_pages > 1 ? (
            <div data-testid="listing-pagination" className="mt-12 flex items-center justify-center gap-2">
              {Array.from({ length: listingsPage.total_pages }, (_, index) => index + 1).map((pageNumber) => {
                const params = new URLSearchParams();
                for (const [key, value] of Object.entries(resolvedSearchParams)) {
                  const normalized = getQueryValue(value);
                  if (normalized) params.set(key, normalized);
                }
                params.set("page", String(pageNumber));

                const isCurrent = pageNumber === listingsPage.page;
                return (
                  <Link
                    key={pageNumber}
                    href={`/listings?${params.toString()}`}
                    className={`flex h-8 min-w-8 items-center justify-center rounded-full px-3 text-sm font-semibold transition ${
                      isCurrent ? "bg-black text-white" : "border border-gray-300 text-[#111] hover:bg-gray-50"
                    }`}
                  >
                    {pageNumber}
                  </Link>
                );
              })}
            </div>
          ) : null}
        </div>
      </main>

      <Footer />
    </div>
  );
}

function getQueryValue(value: string | string[] | undefined) {
  if (Array.isArray(value)) {
    return value[0];
  }
  return value;
}
