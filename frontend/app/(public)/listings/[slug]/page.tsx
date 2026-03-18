import Image from "next/image";

import { getListingBySlug } from "@/features/listings/server/get-listing-by-slug";

export default async function PublicListingDetailPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const listing = await getListingBySlug(slug);
  const image = listing.images[0] ?? null;

  return (
    <main className="min-h-screen px-6 py-10 sm:px-10 lg:px-12">
      <div className="mx-auto grid max-w-6xl gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <section className="relative min-h-96 overflow-hidden rounded-[2rem] border border-[var(--line)] bg-[var(--panel)]">
          {image ? (
            <Image alt={listing.title} fill priority sizes="(min-width: 1024px) 55vw, 100vw" src={image.url} className="object-cover" unoptimized />
          ) : (
            <div className="flex h-full items-center justify-center px-6 text-center text-sm text-[var(--muted)]">No image uploaded yet</div>
          )}
        </section>
        <section className="rounded-[2rem] border border-white/60 bg-[var(--panel)] p-8 shadow-[0_30px_100px_rgba(15,23,42,0.18)] backdrop-blur sm:p-10">
          <p className="text-xs font-semibold uppercase tracking-[0.3em] text-[var(--accent)]">{listing.category?.name ?? "Uncategorized"}</p>
          <h1 className="mt-4 text-4xl font-semibold tracking-[-0.04em] text-[var(--ink)]">{listing.title}</h1>
          <p className="mt-4 text-sm leading-7 text-[var(--muted)]">{listing.description ?? "No public description provided yet."}</p>
          <dl className="mt-8 grid gap-4 text-sm text-[var(--muted)] sm:grid-cols-2">
            <div className="rounded-[1.5rem] border border-[var(--line)] bg-white/72 p-4">
              <dt className="text-xs uppercase tracking-[0.24em]">City</dt>
              <dd className="mt-2 font-medium text-[var(--ink)]">{listing.location_city ?? "Unknown city"}</dd>
            </div>
            <div className="rounded-[1.5rem] border border-[var(--line)] bg-white/72 p-4">
              <dt className="text-xs uppercase tracking-[0.24em]">Status</dt>
              <dd className="mt-2 font-medium text-[var(--ink)]">{listing.status}</dd>
            </div>
          </dl>
        </section>
      </div>
    </main>
  );
}
