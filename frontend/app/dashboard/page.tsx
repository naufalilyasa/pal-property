import Link from "next/link";

import {
  getSellerListings,
  isUnauthenticatedSellerListingsError,
  type SellerListing,
} from "@/lib/api/seller-listings";
import { getRequestCookieHeader } from "@/lib/server/cookies";
import { redirect } from "next/navigation";

function formatPrice(price: number, currency: string) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: currency || "IDR",
    maximumFractionDigits: 0,
  }).format(price);
}

function formatStatus(status: string) {
  return status
    .split("_")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
    .join(" ");
}

function getPrimaryImage(listing: SellerListing) {
  return listing.images.find((image) => image.is_primary) ?? listing.images[0] ?? null;
}

async function loadSellerListings(cookieHeader?: string) {
  try {
    const listingsPage = await getSellerListings({ cookieHeader });

    return {
      status: "success" as const,
      listingsPage,
    };
  } catch (error) {
    if (isUnauthenticatedSellerListingsError(error)) {
      return {
        status: "unauthenticated" as const,
      };
    }

    return {
      status: "error" as const,
    };
  }
}

function SellerState({
  eyebrow,
  title,
  body,
  action,
}: {
  eyebrow: string;
  title: string;
  body: string;
  action?: React.ReactNode;
}) {
  return (
    <section className="rounded-[1.75rem] border border-(--line) bg-white/72 p-8">
      <p
        className="text-xs uppercase tracking-[0.3em] text-(--muted)"
        style={{ fontFamily: "var(--font-mono), monospace" }}
      >
        {eyebrow}
      </p>
      <h2 className="mt-4 text-2xl font-semibold tracking-[-0.03em] text-(--ink)">{title}</h2>
      <p className="mt-3 max-w-2xl text-sm leading-7 text-(--muted) sm:text-base">{body}</p>
      {action ? <div className="mt-6">{action}</div> : null}
    </section>
  );
}

function SellerListingCard({ listing }: { listing: SellerListing }) {
  const image = getPrimaryImage(listing);

  return (
    <article className="grid gap-5 rounded-[1.75rem] border border-(--line) bg-white/80 p-5 sm:grid-cols-[176px_1fr] sm:p-6">
      <div className="overflow-hidden rounded-[1.25rem] border border-(--line) bg-(--panel-strong)">
        {image ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            alt={listing.title}
            className="h-44 w-full object-cover sm:h-full"
            src={image.url}
          />
        ) : (
          <div className="flex h-44 items-center justify-center px-6 text-center text-sm text-(--muted) sm:h-full">
            No image uploaded yet
          </div>
        )}
      </div>

      <div className="flex flex-col gap-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="space-y-2">
            <div className="flex flex-wrap items-center gap-2">
              <span className="rounded-full bg-(--panel-strong) px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-(--accent)">
                {formatStatus(listing.status)}
              </span>
              <span className="text-sm text-(--muted)">{listing.category?.name ?? "Uncategorized"}</span>
            </div>
            <h2 className="text-2xl font-semibold tracking-[-0.03em] text-(--ink)">
              {listing.title}
            </h2>
            <div>
              <Link
                className="inline-flex items-center rounded-full border border-(--line) bg-(--panel) px-4 py-2 text-sm font-semibold text-(--ink) transition hover:border-(--accent) hover:text-(--accent)"
                href={`/dashboard/listings/${listing.id}/edit`}
              >
                Edit listing
              </Link>
            </div>
          </div>

          <p className="text-lg font-semibold text-(--ink)">{formatPrice(listing.price, listing.currency)}</p>
        </div>

        <dl className="grid gap-3 text-sm text-(--muted) sm:grid-cols-2 xl:grid-cols-4">
          <div className="rounded-2xl border border-(--line) bg-(--panel) px-4 py-3">
            <dt className="text-xs uppercase tracking-[0.2em]">Primary image</dt>
            <dd className="mt-2 font-medium text-(--ink)">{image ? "Ready" : "Missing"}</dd>
          </div>
          <div className="rounded-2xl border border-(--line) bg-(--panel) px-4 py-3">
            <dt className="text-xs uppercase tracking-[0.2em]">Views</dt>
            <dd className="mt-2 font-medium text-(--ink)">{listing.view_count}</dd>
          </div>
          <div className="rounded-2xl border border-(--line) bg-(--panel) px-4 py-3">
            <dt className="text-xs uppercase tracking-[0.2em]">City</dt>
            <dd className="mt-2 font-medium text-(--ink)">{listing.location_city ?? "Not set"}</dd>
          </div>
          <div className="rounded-2xl border border-(--line) bg-(--panel) px-4 py-3">
            <dt className="text-xs uppercase tracking-[0.2em]">Updated</dt>
            <dd className="mt-2 font-medium text-(--ink)">
              {new Intl.DateTimeFormat("en", {
                dateStyle: "medium",
              }).format(new Date(listing.updated_at))}
            </dd>
          </div>
        </dl>
      </div>
    </article>
  );
}

export default async function DashboardPage() {
  const listingsResult = await loadSellerListings(await getRequestCookieHeader());

  if (listingsResult.status === "unauthenticated") {
    redirect("/");
  }

  if (listingsResult.status === "error") {
    return (
      <SellerState
        eyebrow="Listings unavailable"
        title="We could not load your listings"
        body="The seller dashboard is available, but the listings service did not respond with usable data. Try again once the API is reachable."
      />
    );
  }

  if (listingsResult.listingsPage.data.length === 0) {
    return (
      <SellerState
        eyebrow="Listings"
        title="No listings yet"
        body="Your seller account is connected, but there are no property records to review yet. Create and publish inventory in the next task flow."
        action={
          <Link
            className="inline-flex items-center rounded-full border border-(--line) bg-(--panel-strong) px-5 py-3 text-sm font-semibold text-(--ink) transition hover:border-(--accent) hover:text-(--accent)"
            href="/dashboard/listings/new"
          >
            Create your first listing
          </Link>
        }
      />
    );
  }

  return (
    <div className="space-y-6">
      <section className="flex justify-end">
        <Link
          className="inline-flex items-center rounded-full bg-(--accent) px-5 py-3 text-sm font-semibold text-white transition hover:opacity-90"
          href="/dashboard/listings/new"
        >
          New listing
        </Link>
      </section>

      <section className="grid gap-4 md:grid-cols-2">
        <div className="rounded-3xl border border-(--line) bg-white/72 p-5">
          <p className="text-xs uppercase tracking-[0.24em] text-(--muted)">Total listings</p>
          <p className="mt-3 text-3xl font-semibold tracking-[-0.04em] text-(--ink)">
            {listingsResult.listingsPage.total}
          </p>
        </div>
        <div className="rounded-3xl border border-(--line) bg-white/72 p-5">
          <p className="text-xs uppercase tracking-[0.24em] text-(--muted)">Current page</p>
          <p className="mt-3 text-3xl font-semibold tracking-[-0.04em] text-(--ink)">
            {listingsResult.listingsPage.page}
          </p>
        </div>
      </section>

      <section className="space-y-5">
        {listingsResult.listingsPage.data.map((listing) => (
          <SellerListingCard key={listing.id} listing={listing} />
        ))}
      </section>
    </div>
  );
}
