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
    limit: getQueryValue(resolvedSearchParams.limit) ?? "10",
    city: getQueryValue(resolvedSearchParams.city),
    category_id: getQueryValue(resolvedSearchParams.category_id),
    price_min: getQueryValue(resolvedSearchParams.price_min),
    price_max: getQueryValue(resolvedSearchParams.price_max),
    status: getQueryValue(resolvedSearchParams.status),
  });

  return (
    <main className="min-h-screen px-6 py-10 sm:px-10 lg:px-12">
      <div className="mx-auto max-w-6xl space-y-6">
        <section className="rounded-[2rem] border border-white/60 bg-[var(--panel)] p-8 shadow-[0_30px_100px_rgba(15,23,42,0.18)] backdrop-blur sm:p-10">
          <p className="text-xs font-semibold uppercase tracking-[0.3em] text-[var(--accent)]">Public listings</p>
          <h1 className="mt-4 text-4xl font-semibold tracking-[-0.04em] text-[var(--ink)]">Browse available PAL Property inventory</h1>
          <p className="mt-4 text-sm leading-7 text-[var(--muted)]">Server-rendered listing results stay aligned with backend filters while client-side filter helpers can hydrate independently.</p>
        </section>
        <ListingFilters />
        <section className="grid gap-5 lg:grid-cols-2 xl:grid-cols-3">
          {listingsPage.data.map((listing) => (
            <div data-testid="listing-card" key={listing.id}>
              <ListingCard href={`/listings/${listing.slug}`} listing={listing} />
            </div>
          ))}
        </section>
        <section className="rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-5" data-testid="listing-pagination">
          Page {listingsPage.page} of {Math.max(listingsPage.total_pages, 1)} · showing {listingsPage.data.length} listing(s)
        </section>
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
