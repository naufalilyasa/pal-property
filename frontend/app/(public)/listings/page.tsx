import Link from "next/link";

import { ListingFilters } from "@/features/listings/components/listing-filters";
import { ListingCard } from "@/features/listings/components/listing-card";
import { getListings } from "@/features/listings/server/get-listings";

export default async function PublicListingsPage({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
}) {
  const resolvedSearchParams = (await searchParams) ?? {};
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

  return (
    <main className="min-h-screen bg-[#f5f4ef] px-3 py-4 sm:px-5 lg:px-6">
      <div className="mx-auto max-w-[1580px] space-y-3">
        <section className="rounded-[20px] border border-[#e8e4db] bg-white px-4 py-3 shadow-[0_8px_32px_rgba(17,17,17,0.045)] sm:px-5">
          <div className="flex flex-col gap-3 border-b border-[#ede9e0] pb-3 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex flex-wrap items-center gap-x-6 gap-y-2">
              <Link className="text-[1.45rem] font-black uppercase tracking-[-0.08em] text-[#111]" href="/listings">
                PAL FIND
              </Link>
              <nav className="flex flex-wrap items-center gap-4 text-[12px] font-medium text-[#5e5a54]">
                <span>Search</span>
                <span>Agents</span>
                <span>Join</span>
                <span>Resources</span>
                <span>About</span>
              </nav>
            </div>
            <div className="flex items-center gap-2 self-start lg:self-auto">
              <Link className="rounded-md bg-[#111] px-4 py-2 text-[12px] font-semibold uppercase tracking-[0.16em] text-white" href="/login">
                Sign in
              </Link>
            </div>
          </div>

          <div className="flex flex-col gap-3 pt-3 xl:flex-row xl:items-start xl:justify-between">
            <div className="min-w-0 flex-1">
              <ListingFilters
                key={JSON.stringify(filterValues)}
                total={listingsPage.total}
                visibleCount={listingsPage.data.length}
                values={filterValues}
              />
            </div>
            <div className="flex items-center gap-2 xl:pl-3">
              <button aria-disabled="true" className="cursor-default rounded-md bg-[#111] px-4 py-2 text-[12px] font-semibold uppercase tracking-[0.16em] text-white opacity-85" type="button">
                Map Preview
              </button>
              <button aria-disabled="true" className="cursor-default rounded-md border border-[#d7d2c9] bg-white px-4 py-2 text-[12px] font-semibold uppercase tracking-[0.16em] text-[#111] opacity-60" type="button">
                List Preview
              </button>
            </div>
          </div>
        </section>

        <div className="grid gap-4 xl:grid-cols-[minmax(340px,0.98fr)_minmax(0,1.02fr)] xl:items-start">
          <section className="order-1 xl:sticky xl:top-4" data-testid="listing-map-panel">
            <div className="overflow-hidden rounded-[18px] border border-[#ddd9cf] bg-white shadow-[0_10px_30px_rgba(17,17,17,0.04)]">
              <div className="flex items-center justify-between border-b border-[#ece8df] px-3 py-2 text-[11px] font-medium uppercase tracking-[0.16em] text-[#6e6a64]">
                <span>Map view</span>
                <span>Jakarta metro</span>
              </div>
              <div className="h-[320px] sm:h-[440px] xl:h-[calc(100vh-12rem)] xl:min-h-[760px]">
                <iframe
                  className="h-full w-full border-0"
                  loading="lazy"
                  referrerPolicy="no-referrer-when-downgrade"
                  src="https://www.openstreetmap.org/export/embed.html?bbox=106.5608%2C-6.3754%2C107.0217%2C-5.9949&layer=mapnik"
                  title="Property search map"
                />
              </div>
            </div>
          </section>

          <div className="order-2 space-y-4">
            <section className="flex flex-col gap-2 border-b border-[#e8e4db] pb-2 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p className="text-[12px] font-medium text-[#2d2a25]">{listingsPage.total.toLocaleString("en-US")} Results</p>
                <p className="text-[10px] uppercase tracking-[0.14em] text-[#7e7b75]">Showing {listingsPage.data.length} on page {listingsPage.page}</p>
              </div>
              <div className="flex items-center gap-4 text-[10px] font-medium uppercase tracking-[0.14em] text-[#6e6a64]">
                <span>Sort</span>
                <span className="text-[#111]">Newest</span>
                <span className="rounded-full bg-[#f1efe9] px-2 py-1 text-[#7e7b75]">Preview</span>
              </div>
            </section>

            <section className="grid gap-x-5 gap-y-6 md:grid-cols-2">
              {listingsPage.data.length === 0 ? (
                <div className="col-span-full rounded-sm border border-dashed border-[#d8d5cf] bg-white p-10 text-center">
                  <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-[#7e7b75]">No results</p>
                  <h3 className="mt-3 text-xl font-semibold tracking-[-0.03em] text-[#111]">Try broadening the search</h3>
                  <p className="mt-2 text-sm text-[#6d6a64]">Remove one or two filters, increase the price range, or search by a broader city name to surface more inventory.</p>
                </div>
              ) : null}

              {listingsPage.data.map((listing) => (
                <div data-testid="listing-card" key={listing.id}>
                  <ListingCard href={`/listings/${listing.slug}`} listing={listing} />
                </div>
              ))}
            </section>

            <section className="flex items-center justify-center gap-2 pt-1" data-testid="listing-pagination">
              <span className="flex h-7 w-7 items-center justify-center rounded-full bg-[#111] text-[11px] font-semibold text-white">{listingsPage.page}</span>
              <span className="text-[11px] text-[#6e6a64]">2</span>
              <span className="text-[11px] text-[#6e6a64]">3</span>
              <span className="text-[11px] text-[#6e6a64]">...</span>
              <span className="text-[11px] text-[#6e6a64]">{Math.max(listingsPage.total_pages, 1)}</span>
              <span className="ml-1 flex h-7 w-7 items-center justify-center rounded-full border border-[#d8d5cf] text-[11px] text-[#111]">&gt;</span>
            </section>

            <section className="rounded-[22px] bg-[#181818] px-5 py-8 text-white sm:px-8 sm:py-10">
              <div className="grid gap-8 lg:grid-cols-[1.35fr_0.65fr]">
                <div>
                  <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-white/55">Newsletter</p>
                  <h3 className="mt-3 text-2xl font-semibold tracking-[-0.04em]">Subscribe to our listing updates</h3>
                  <div className="mt-5 flex max-w-xl items-center gap-3 border-b border-white/20 pb-3 text-sm text-white/70">
                    <span>Enter email address</span>
                    <span className="ml-auto">&gt;</span>
                  </div>
                  <p className="mt-2 text-[10px] font-medium uppercase tracking-[0.16em] text-white/40">Preview only</p>
                </div>
                <div className="grid grid-cols-2 gap-6 text-sm text-white/75">
                  <div>
                    <p className="font-semibold text-white">Search</p>
                    <p className="mt-2">Agents</p>
                    <p className="mt-1">Join</p>
                    <p className="mt-1">About Us</p>
                  </div>
                  <div>
                    <p className="font-semibold text-white">Follow</p>
                    <p className="mt-2">Facebook</p>
                    <p className="mt-1">Instagram</p>
                    <p className="mt-1">Youtube</p>
                  </div>
                </div>
              </div>
            </section>
          </div>
        </div>
      </div>
    </main>
  );
}

function getQueryValue(value: string | string[] | undefined) {
  if (Array.isArray(value)) {
    return value[0];
  }

  return value;
}
