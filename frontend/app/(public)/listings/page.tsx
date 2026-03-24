import Link from "next/link";

import { ListingFilters } from "@/features/listings/components/listing-filters";
import { ListingCard } from "@/features/listings/components/listing-card";
import { getListings } from "@/features/listings/server/get-listings";
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
    city: getQueryValue(resolvedSearchParams.city),
    category_id: getQueryValue(resolvedSearchParams.category_id),
    price_min: getQueryValue(resolvedSearchParams.price_min),
    price_max: getQueryValue(resolvedSearchParams.price_max),
    status: getQueryValue(resolvedSearchParams.status),
    limit: getQueryValue(resolvedSearchParams.limit) ?? "12",
  };

  const listingsPage = await getListings({
    page: getQueryValue(resolvedSearchParams.page) ?? "1",
    limit: filterValues.limit,
    city: filterValues.city,
    category_id: filterValues.category_id,
    price_min: filterValues.price_min,
    price_max: filterValues.price_max,
    status: filterValues.status,
  });

  // Inject dummy data to fill the grid if we don't have enough real listings
  const mockListingsCount = Math.max(0, 12 - listingsPage.data.length);
  const dummyListings = Array.from({ length: mockListingsCount }).map(
    (_, i) =>
      ({
        id: `dummy-${i}`,
        title: `Premium Property ${i}`,
        slug: `dummy-${i}`,
        price: 0, // Fallback triggering in ListingCard
        currency: "USD",
        status: "active",
        view_count: 0,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        images: [],
        specifications: {},
      }) as any, // Cast to any to satisfy the complex ListingRecord type simply
  );

  const displayListings = [...listingsPage.data, ...dummyListings];
  const displayTotal = listingsPage.total + mockListingsCount;

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
          <div className="hidden w-[45%] border-r border-gray-200 lg:block">
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
                  Try broadening the search
                </h3>
                <p className="mt-2 text-sm text-gray-600">
                  Remove filters or search a broader area.
                </p>
              </div>
            ) : null}

            {displayListings.map((listing) => (
              <ListingCard
                key={listing.id}
                href={`/listings/${listing.slug}`}
                listing={listing}
              />
            ))}
          </div>

          {/* Pagination */}
          <div className="mt-12 flex items-center justify-center gap-2">
            <button className="flex h-8 w-8 items-center justify-center rounded-full border border-gray-300 text-sm font-semibold transition hover:bg-gray-50">
              &lt;
            </button>
            <span className="flex h-8 w-8 items-center justify-center rounded-full bg-black text-sm font-semibold text-white">
              1
            </span>
            <span className="flex h-8 w-8 cursor-pointer items-center justify-center rounded-full text-sm font-semibold hover:bg-gray-50">
              2
            </span>
            <span className="flex h-8 w-8 cursor-pointer items-center justify-center rounded-full text-sm font-semibold hover:bg-gray-50">
              3
            </span>
            <span className="flex h-8 w-8 items-center justify-center text-sm font-semibold">
              ...
            </span>
            <span className="flex h-8 w-8 cursor-pointer items-center justify-center rounded-full text-sm font-semibold hover:bg-gray-50">
              10
            </span>
            <button className="flex h-8 w-8 items-center justify-center rounded-full border border-gray-300 text-sm font-semibold transition hover:bg-gray-50">
              &gt;
            </button>
          </div>
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
