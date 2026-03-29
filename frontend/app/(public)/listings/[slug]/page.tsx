import Image from "next/image";
import { getOptionalUser } from "@/features/auth/server/current-user";
import { getListingBySlug } from "@/features/listings/server/get-listing-by-slug";
import { SaveListingButton } from "@/features/saved-listings/components/save-listing-button";
import { getSavedListingIdsForListings } from "@/features/saved-listings/server/get-saved-listing-ids";
import { TopNav } from "@/features/listings/components/top-nav";
import { Footer } from "@/features/listings/components/footer";
import type { ListingSpecifications, ListingRecord } from "@/lib/api/listing-form";

function formatPrice(price: number, currency: string) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currency || "USD",
    maximumFractionDigits: 0,
  }).format(price);
}

function parseSpecifications(specifications: unknown): Partial<ListingSpecifications> {
  if (!specifications || typeof specifications !== "object") return {};
  const candidate = specifications as Record<string, unknown>;
  const tonum = (val: unknown) => (typeof val === "number" && Number.isFinite(val) ? val : undefined);
  return {
    bedrooms: tonum(candidate.bedrooms),
    bathrooms: tonum(candidate.bathrooms),
    building_area_sqm: tonum(candidate.building_area_sqm),
  };
}

function formatArea(value?: number | null) {
  if (!value || value <= 0) return null;
  return `${value} sqm`;
}

function formatList(value?: string[] | null) {
  return value && value.length > 0 ? value.join(", ") : null;
}

export default async function PublicListingDetailPage({ params }: { params: Promise<{ slug: string }> }) {
  const [{ slug }, user] = await Promise.all([params, getOptionalUser()]);

  let listing: ListingRecord | null = null;
  try {
    if (!slug.includes("dummy")) {
      listing = await getListingBySlug(slug);
    }
  } catch {
    // Graceful fallback for failed fetches to continue displaying UI
  }

  const initialSaved = user && listing?.id
    ? (await getSavedListingIdsForListings([listing.id])).listingIds.includes(listing.id)
    : false;

  // DUMMY DATA FALLBACK
  const dummy = {
    address: "311 Glen Avenue, Scotia, NY, 12302",
    status: "Active",
    category: "Condo",
    price: 230000,
    currency: "USD",
    beds: 3,
    baths: 2,
    sqft: 1200,
    description:
      "Welcome to 311 Glen, an updated colonial, with 3 bedrooms and 2 full bathrooms on an oversized lot with an additional office and family room in the Scotia Glenville School District. This home has been lovingly cared for over a decade. The living room & formal dining room have updated flooring. The updated eat-in kitchen includes space for a breakfast table. The first floor primary bedroom and 1 of 2 full modern baths include laundry on the first floor for additional convenience. The generous 4 season front porch has brand new windows, ideal for additional storage. Other updates include an updated electric panel, newly insulated attic, vinyl replacement windows, updated water heater and refrigerator. This home is close to public transit, public parks with nature trails all within the award-winning Scotia Glenville Schools. Check out the Showcase Video or 3d the camp scale pre and digital side of this home.",
    images: [
      "https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&q=80&w=1200",
      "https://images.unsplash.com/photo-1600596542815-ffad4c1539a9?auto=format&fit=crop&q=80&w=1200",
      "https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&q=80&w=800",
      "https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&q=80&w=800",
      "https://images.unsplash.com/photo-1600566753086-00f18efc2291?auto=format&fit=crop&q=80&w=1200",
      "https://images.unsplash.com/photo-1600573472550-8090b5e0745e?auto=format&fit=crop&q=80&w=800",
      "https://images.unsplash.com/photo-1600047509807-ba8f99d2cd58?auto=format&fit=crop&q=80&w=800",
    ],
    details: [
      { label: "Property type", value: "House" },
      { label: "Year built", value: "1920" },
      { label: "Lot size", value: "-" },
      { label: "Central A/C", value: "Yes" },
      { label: "Heating", value: "Yes" },
      { label: "Garage", value: "No" },
    ],
  };

  const address = listing?.title || dummy.address;
  const status = listing?.status && listing.status !== "active" ? listing.status : dummy.status;
  const category = listing?.category?.name || dummy.category;
  const priceNum = listing?.price && listing.price > 0 ? listing.price : dummy.price;
  const price = formatPrice(priceNum, listing?.currency || dummy.currency);

  const specs = parseSpecifications(listing?.specifications);
  const beds = listing?.bedroom_count ?? specs.bedrooms ?? dummy.beds;
  const baths = listing?.bathroom_count ?? specs.bathrooms ?? dummy.baths;
  const buildingAreaSqm = listing?.building_area_sqm ?? specs.building_area_sqm;
  const sqft = buildingAreaSqm
    ? Math.round(buildingAreaSqm * 10.7639)
    : dummy.sqft;

  const descriptionText = listing?.description || dummy.description;
  const details = [
    { label: "Transaction", value: listing?.transaction_type ?? "sale" },
    { label: "Province", value: listing?.location_province ?? null },
    { label: "Certificate", value: listing?.certificate_type ?? null },
    { label: "Condition", value: listing?.condition ?? null },
    { label: "Furnishing", value: listing?.furnishing ?? null },
    { label: "Electrical", value: listing?.electrical_power_va ? `${listing.electrical_power_va} VA` : null },
    { label: "Facing", value: listing?.facing_direction ?? null },
    { label: "Year built", value: listing?.year_built ? String(listing.year_built) : dummy.details[1].value },
    { label: "Land area", value: formatArea(listing?.land_area_sqm ?? specs.land_area_sqm) ?? dummy.details[2].value },
    { label: "Building area", value: formatArea(buildingAreaSqm) },
    { label: "Floors", value: listing?.floor_count ? String(listing.floor_count) : null },
    { label: "Carport", value: listing?.carport_capacity ? String(listing.carport_capacity) : dummy.details[5].value },
    { label: "Negotiable", value: listing?.is_negotiable ? "Yes" : null },
    { label: "Facilities", value: formatList(listing?.facilities) },
    { label: "Special offers", value: formatList(listing?.special_offers) },
  ].filter((detail): detail is { label: string; value: string } => Boolean(detail.value));

  // Ensure we have enough images for the grid layout visually
  let images = listing?.images?.length && listing.images.length > 0
    ? listing.images.map((i) => i.url)
    : [];
  if (images.length < dummy.images.length) {
    images = images.concat(dummy.images.slice(images.length));
  }

  return (
    <div className="flex min-h-screen flex-col bg-background font-sans text-foreground">
      <TopNav />

      <main className="flex-1">
        <div className="mx-auto w-full max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
          
          {/* Header Section */}
          <div className="flex flex-col justify-between gap-6 pb-6 md:flex-row md:items-start">
            <div>
              <h1 className="text-3xl font-light tracking-tight text-foreground sm:text-4xl">
                {address}
              </h1>
              <div className="mt-3 flex items-center gap-2">
                <span className="rounded-full bg-green-100 px-3 py-0.5 text-xs font-semibold text-green-800 uppercase tracking-widest">{status}</span>
                <span className="rounded-full border border-border px-3 py-0.5 text-xs font-semibold uppercase tracking-widest text-muted-foreground">{category}</span>
              </div>
            </div>

            <div className="flex flex-col items-end justify-between gap-10">
              <button type="button" aria-label="More options" className="hidden h-10 w-10 items-center justify-center rounded-full border border-border bg-white text-foreground shadow-sm transition-colors hover:bg-slate-50 md:flex">
                <svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M11 17h2"/><path d="M11 12h2"/><path d="M11 7h2"/></svg>
              </button>

              <div className="flex flex-row items-center gap-2">
                <button type="button" className="inline-flex h-10 items-center justify-center rounded-full bg-slate-900 px-6 text-sm font-semibold text-white shadow transition-colors hover:bg-slate-900/90">
                  Schedule a tour
                </button>
                {listing?.id ? (
                  <SaveListingButton
                    initialSaved={initialSaved}
                    listingId={listing.id}
                    variant="icon"
                  />
                ) : (
                  <button type="button" aria-label="Save" className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-border bg-white text-foreground shadow-sm transition-colors hover:bg-slate-50">
                    <svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z"/></svg>
                  </button>
                )}
                <button type="button" aria-label="Share" className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-border bg-white text-foreground shadow-sm transition-colors hover:bg-slate-50">
                  <svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"/><polyline points="16 6 12 2 8 6"/><line x1="12" x2="12" y1="2" y2="15"/></svg>
                </button>
              </div>
            </div>
          </div>

          {/* Hero Image */}
          <div className="relative mt-2 aspect-video w-full overflow-hidden sm:aspect-[2.35/1]">
            <Image
              alt={address}
              fill
              priority
              src={images[0]}
              className="object-cover"
              unoptimized
            />
          </div>

          {/* Price and Specs */}
          <div className="mt-6 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
            <div className="text-4xl font-semibold tracking-tight text-foreground sm:text-5xl">
              {price}
            </div>
            
            <div className="flex gap-10">
              <div className="flex flex-col">
                <span className="text-xl font-semibold">{beds}</span>
                <span className="text-sm font-medium text-muted-foreground uppercase tracking-widest">Beds</span>
              </div>
              <div className="flex flex-col">
                <span className="text-xl font-semibold">{baths}</span>
                <span className="text-sm font-medium text-muted-foreground uppercase tracking-widest">Baths</span>
              </div>
              <div className="flex flex-col">
                <span className="text-xl font-semibold">{sqft}</span>
                <span className="text-sm font-medium text-muted-foreground uppercase tracking-widest">SqFt</span>
              </div>
            </div>
          </div>

          <hr className="my-12 border-border" />

          {/* About Section */}
          <section className="grid gap-8 lg:grid-cols-[250px_1fr]">
            <div>
              <h2 className="text-2xl font-light text-foreground leading-[1.1]">
                About <br/><span className="text-muted-foreground">this home</span>
              </h2>
            </div>
            <div>
              <div className="mb-6 flex flex-wrap gap-2">
                <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-foreground shadow-sm">{category}</span>
                <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-foreground shadow-sm">{listing?.transaction_type ?? "sale"}</span>
                {listing?.condition ? <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-foreground shadow-sm">{listing.condition}</span> : null}
                {listing?.furnishing ? <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-foreground shadow-sm">{listing.furnishing}</span> : null}
                {listing?.year_built ? <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-foreground shadow-sm">Year built {listing.year_built}</span> : null}
              </div>
              <p className="text-base leading-relaxed text-foreground">
                {descriptionText}
              </p>
            </div>
          </section>

          {/* Dynamic Image Gallery block */}
          <section className="mt-16 flex flex-col gap-4">
            <div className="relative aspect-16/6 w-full overflow-hidden bg-muted">
               <Image alt="gallery 1" fill src={images[1]} className="object-cover" unoptimized/>
            </div>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
               <div className="relative aspect-4/3 w-full overflow-hidden bg-muted">
                 <Image alt={`${address} gallery image 2`} fill src={images[2]} className="object-cover" unoptimized/>
               </div>
               <div className="grid grid-rows-2 gap-4">
                 <div className="relative w-full h-full overflow-hidden bg-muted">
                    <Image alt={`${address} gallery image 3`} fill src={images[3]} className="object-cover" unoptimized/>
                 </div>
                 <div className="relative w-full h-full overflow-hidden bg-muted">
                    <Image alt={`${address} gallery image 4`} fill src={images[4]} className="object-cover" unoptimized/>
                 </div>
               </div>
            </div>
            <div className="relative aspect-16/6 w-full overflow-hidden bg-muted">
               <Image alt={images[5] ? `${address} gallery image 5` : `${address} gallery image`} fill src={images[5]} className="object-cover" unoptimized/>
            </div>
          </section>

          <hr className="my-16 border-border" />

          {/* Other Details Section */}
          <section className="grid gap-8 lg:grid-cols-[250px_1fr]">
            <div>
              <h2 className="text-2xl font-light text-foreground leading-[1.1]">
                Other <br/><span className="text-muted-foreground">Details</span>
              </h2>
            </div>
            <div>
              <dl className="grid grid-cols-2 gap-y-10 sm:grid-cols-3">
                {details.map((detail) => (
                  <div key={detail.label} className="flex flex-col">
                    <dt className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground">{detail.label}</dt>
                    <dd className="mt-1 text-sm font-semibold text-foreground">{detail.value}</dd>
                  </div>
                ))}
              </dl>
            </div>
          </section>

          {/* Carousel footer dummy section */}
          <section className="mt-16 flex flex-col">
             <div className="grid grid-cols-2 gap-4 h-[240px] w-full overflow-hidden md:h-[400px]">
                <div className="relative w-full h-full overflow-hidden bg-muted">
                   <Image alt={`${address} gallery image left`} fill src={images[0]} className="object-cover" unoptimized/>
                </div>
                <div className="relative w-full h-full overflow-hidden bg-muted">
                   <Image alt={`${address} gallery image right`} fill src={images[5]} className="object-cover" unoptimized/>
                </div>
             </div>
             <div className="mt-6 flex items-center justify-between">
               <div className="text-sm font-semibold">14 <span className="text-muted-foreground">/ 21</span></div>
               <div className="h-1 hidden w-64 items-center bg-muted md:flex">
                 <div className="h-full w-2/3 bg-slate-900"></div>
               </div>
               <div className="flex gap-2">
                 <button type="button" className="flex h-8 w-8 items-center justify-center rounded-full bg-slate-900 text-white shadow hover:opacity-90">
                    &lt;
                 </button>
                 <button type="button" className="flex h-8 w-8 items-center justify-center rounded-full bg-slate-900 text-white shadow hover:opacity-90">
                    &gt;
                 </button>
               </div>
             </div>
          </section>

          {/* Map */}
          <section className="mt-16">
            <div className="h-[300px] w-full overflow-hidden rounded-xl border border-border sm:h-[400px]">
              <iframe
                className="h-full w-full border-0 grayscale-10"
                loading="lazy"
                referrerPolicy="no-referrer-when-downgrade"
                src={`https://www.openstreetmap.org/export/embed.html?bbox=106.5608%2C-6.3754%2C107.0217%2C-5.9949&layer=mapnik&marker=-6.2088%2C106.8456`}
                title="Property search map location"
              />
            </div>
          </section>

        </div>
      </main>

      <Footer />
    </div>
  );
}
