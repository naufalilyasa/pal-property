import { ListingFilters } from "@/features/listings/components/listing-filters";
import { ListingCard } from "@/features/listings/components/listing-card";
import { getListings } from "@/features/listings/server/get-listings";

export default async function PublicListingsPage({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
}) {
  const resolvedSearchParams = (await searchParams) ?? {};
  const listingsPage = await getListings({
    page: getQueryValue(resolvedSearchParams.page) ?? "1",
    limit: getQueryValue(resolvedSearchParams.limit) ?? "12",
    city: getQueryValue(resolvedSearchParams.city),
    category_id: getQueryValue(resolvedSearchParams.category_id),
    price_min: getQueryValue(resolvedSearchParams.price_min),
    price_max: getQueryValue(resolvedSearchParams.price_max),
    status: getQueryValue(resolvedSearchParams.status),
  });

  return (
    <main className="min-h-screen bg-[#f8f8f5] px-4 py-5 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-[1560px] space-y-4">
        <section className="rounded-2xl border border-[#ece9e2] bg-white px-4 py-4 shadow-[0_8px_30px_rgba(17,17,17,0.04)] sm:px-5">
          <div className="flex flex-col gap-3 border-b border-[#ece9e2] pb-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[#7e7b75]">Search</p>
              <h1 className="mt-1 text-2xl font-semibold tracking-[-0.04em] text-[#111]">Property search</h1>
            </div>
            <div className="flex items-center gap-2">
              <button className="rounded-md bg-[#111] px-4 py-2 text-[12px] font-semibold uppercase tracking-[0.16em] text-white" type="button">Map</button>
              <button className="rounded-md border border-[#dcd9d2] bg-white px-4 py-2 text-[12px] font-semibold uppercase tracking-[0.16em] text-[#111]" type="button">List</button>
            </div>
          </div>

          <div className="pt-3">
            <ListingFilters
              total={listingsPage.total}
              visibleCount={listingsPage.data.length}
              values={{
                city: getQueryValue(resolvedSearchParams.city),
                category_id: getQueryValue(resolvedSearchParams.category_id),
                price_min: getQueryValue(resolvedSearchParams.price_min),
                price_max: getQueryValue(resolvedSearchParams.price_max),
                status: getQueryValue(resolvedSearchParams.status),
                limit: getQueryValue(resolvedSearchParams.limit) ?? "12",
              }}
            />
          </div>
        </section>

        <div className="grid gap-5 xl:grid-cols-[minmax(320px,0.95fr)_minmax(0,1.05fr)] xl:items-start">
          <section className="xl:sticky xl:top-4">
            <div className="overflow-hidden rounded-sm border border-[#ddd] bg-white">
              <div className="flex items-center justify-between border-b border-[#ece9e2] px-3 py-2 text-[11px] font-medium uppercase tracking-[0.16em] text-[#6e6a64]">
                <span>Map view</span>
                <span>Jakarta metro</span>
              </div>
              <div className="h-[360px] sm:h-[520px] xl:h-[760px]">
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

          <div className="space-y-4">
            <section className="flex flex-col gap-2 border-b border-[#e9e6df] pb-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p className="text-[12px] font-medium text-[#2d2a25]">{listingsPage.total.toLocaleString("en-US")} Results</p>
                <p className="text-[11px] text-[#7e7b75]">Showing {listingsPage.data.length} listings on page {listingsPage.page}</p>
              </div>
              <div className="flex items-center gap-3 text-[11px] font-medium uppercase tracking-[0.16em] text-[#6e6a64]">
                <span>Sort</span>
                <span className="text-[#111]">Newest</span>
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

            <section className="flex items-center justify-center gap-2 pt-2" data-testid="listing-pagination">
              <span className="flex h-7 w-7 items-center justify-center rounded-full bg-[#111] text-[11px] font-semibold text-white">{listingsPage.page}</span>
              <span className="text-[11px] text-[#6e6a64]">of {Math.max(listingsPage.total_pages, 1)}</span>
              <span className="ml-2 text-[11px] text-[#6e6a64]">{listingsPage.data.length} shown</span>
              <span className="ml-2 flex h-7 w-7 items-center justify-center rounded-full border border-[#d8d5cf] text-[11px] text-[#111]">&gt;</span>
            </section>

            <section className="rounded-2xl bg-[#181818] px-5 py-8 text-white sm:px-8 sm:py-10">
              <div className="grid gap-8 lg:grid-cols-[1.3fr_0.7fr]">
                <div>
                  <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-white/55">Newsletter</p>
                  <h3 className="mt-3 text-2xl font-semibold tracking-[-0.04em]">Subscribe to our listing updates</h3>
                  <div className="mt-5 flex max-w-xl items-center gap-3 border-b border-white/20 pb-3 text-sm text-white/70">
                    <span>Enter address</span>
                    <span className="ml-auto">&gt;</span>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4 text-sm text-white/75">
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
