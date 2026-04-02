import { getOptionalUser } from "@/features/auth/server/current-user";
import { getListingBySlug } from "@/features/listings/server/get-listing-by-slug";
import { SaveListingButton } from "@/features/saved-listings/components/save-listing-button";
import { getSavedListingIdsForListings } from "@/features/saved-listings/server/get-saved-listing-ids";
import { TopNav } from "@/features/listings/components/top-nav";
import { Footer } from "@/features/listings/components/footer";
import { ListingDetailGallery } from "@/features/listings/components/listing-detail-gallery";
import type { ListingSpecifications, ListingRecord } from "@/lib/api/listing-form";

function formatPrice(price: number, currency: string) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currency || "IDR",
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
    land_area_sqm: tonum(candidate.land_area_sqm),
  };
}

function formatArea(value?: number | null) {
  if (!value || value <= 0) return null;
  return `${value} m²`;
}

function formatList(value?: string[] | null) {
  return value && value.length > 0 ? value.join(", ") : null;
}

function formatVideoDuration(seconds?: number | null) {
  if (!seconds || seconds <= 0) return null;

  const minutes = Math.floor(seconds / 60);
  const remainder = seconds % 60;

  if (minutes === 0) return `${remainder} dtk`;
  if (remainder === 0) return `${minutes} mnt`;
  return `${minutes} mnt ${remainder} dtk`;
}

function formatTransactionType(value?: string | null) {
  if (value === "rent") return "Sewa";
  if (value === "sale") return "Jual";
  return value ?? null;
}

function formatStatus(value?: string | null) {
  if (!value) return null;

  const labels: Record<string, string> = {
    active: "Aktif",
    draft: "Draft",
    sold: "Terjual",
    rented: "Tersewa",
    archived: "Arsip",
  };

  return labels[value] ?? value;
}

function formatCondition(value?: string | null) {
  if (value === "new") return "Properti Baru";
  if (value === "second") return "Properti Second";
  return value ?? null;
}

function formatFurnishing(value?: string | null) {
  if (value === "furnished") return "Furnished";
  if (value === "semi") return "Semi Furnished";
  if (value === "unfurnished") return "Unfurnished";
  return value ?? null;
}

function formatBoolean(value?: boolean | null) {
  if (value == null) return null;
  return value ? "Ya" : "Tidak";
}

function buildMapEmbedUrl(latitude?: number | null, longitude?: number | null) {
  if (latitude == null || longitude == null) {
    return null;
  }

  const latOffset = 0.01;
  const lngOffset = 0.01;
  const bbox = [longitude - lngOffset, latitude - latOffset, longitude + lngOffset, latitude + latOffset]
    .map((value) => value.toFixed(6))
    .join(",");

  return `https://www.openstreetmap.org/export/embed.html?bbox=${bbox}&layer=mapnik&marker=${latitude.toFixed(6)}%2C${longitude.toFixed(6)}`;
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
    status: "Aktif",
    category: "Kondominium",
    price: 230000,
    currency: "IDR",
    beds: 3,
    baths: 2,
    buildingAreaSqm: 111,
    description:
      "Hunian ini menawarkan tata ruang yang nyaman untuk keluarga kecil maupun pasangan muda. Area utama terasa terang, sirkulasi udara baik, dan akses ke fasilitas sekitar cukup praktis untuk kebutuhan harian.",
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
      { label: "Tipe properti", value: "Rumah" },
      { label: "Tahun dibangun", value: "1920" },
      { label: "Luas tanah", value: "-" },
      { label: "AC sentral", value: "Ya" },
      { label: "Pemanas", value: "Ya" },
      { label: "Garasi", value: "Tidak" },
    ],
  };

  const address = listing?.title || dummy.address;
  const status = formatStatus(listing?.status) ?? dummy.status;
  const category = listing?.category?.name || dummy.category;
  const priceNum = listing?.price && listing.price > 0 ? listing.price : dummy.price;
  const price = formatPrice(priceNum, listing?.currency || dummy.currency);

  const specs = parseSpecifications(listing?.specifications);
  const beds = listing?.bedroom_count ?? specs.bedrooms ?? dummy.beds;
  const baths = listing?.bathroom_count ?? specs.bathrooms ?? dummy.baths;
  const buildingAreaSqm = listing?.building_area_sqm ?? specs.building_area_sqm;
  const buildingAreaText = formatArea(buildingAreaSqm) ?? formatArea(dummy.buildingAreaSqm);
  const landAreaText = formatArea(listing?.land_area_sqm ?? specs.land_area_sqm) ?? dummy.details[2].value;
  const locationLine = [listing?.location_village, listing?.location_district, listing?.location_city, listing?.location_province]
    .filter(Boolean)
    .join(", ");
  const mapEmbedUrl = buildMapEmbedUrl(listing?.latitude, listing?.longitude);

  const descriptionText = listing?.description || dummy.description;
  const listingVideo = listing?.video ?? null;
  const details = [
    { label: "Tipe transaksi", value: formatTransactionType(listing?.transaction_type) ?? "Jual" },
    { label: "Provinsi", value: listing?.location_province ?? null },
    { label: "Kota / Kabupaten", value: listing?.location_city ?? null },
    { label: "Kecamatan", value: listing?.location_district ?? null },
    { label: "Kelurahan / Desa", value: listing?.location_village ?? null },
    { label: "Sertifikat", value: listing?.certificate_type ?? null },
    { label: "Kondisi", value: formatCondition(listing?.condition) },
    { label: "Perabotan", value: formatFurnishing(listing?.furnishing) },
    { label: "Daya listrik", value: listing?.electrical_power_va ? `${listing.electrical_power_va} VA` : null },
    { label: "Hadap", value: listing?.facing_direction ?? null },
    { label: "Tahun dibangun", value: listing?.year_built ? String(listing.year_built) : dummy.details[1].value },
    { label: "Luas tanah", value: landAreaText },
    { label: "Luas bangunan", value: buildingAreaText },
    { label: "Jumlah lantai", value: listing?.floor_count ? String(listing.floor_count) : null },
    { label: "Kapasitas carport", value: listing?.carport_capacity ? String(listing.carport_capacity) : dummy.details[5].value },
    { label: "Bisa nego", value: formatBoolean(listing?.is_negotiable) },
    { label: "Fasilitas", value: formatList(listing?.facilities) },
    { label: "Promo khusus", value: formatList(listing?.special_offers) },
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
        <div className="mx-auto w-full max-w-[1440px] px-4 py-8 sm:px-6 lg:px-8">
          
          {/* Header Section */}
          <div className="flex flex-col justify-between gap-6 pb-6 md:flex-row md:items-start">
            <div className="max-w-3xl">
              <h1 className="text-3xl font-light tracking-tight text-foreground sm:text-4xl">
                {address}
              </h1>
              <div className="mt-3 flex items-center gap-2">
                <span className="rounded-full bg-green-100 px-3 py-0.5 text-xs font-semibold text-green-800 uppercase tracking-widest">{status}</span>
                <span className="rounded-full border border-border px-3 py-0.5 text-xs font-semibold uppercase tracking-widest text-muted-foreground">{category}</span>
              </div>
              {locationLine ? <p className="mt-4 text-sm leading-7 text-muted-foreground">{locationLine}</p> : null}
            </div>

            <div className="flex flex-col items-end justify-between gap-10">
              <button type="button" aria-label="More options" className="hidden h-10 w-10 items-center justify-center rounded-full border border-border bg-white text-foreground shadow-sm transition-colors hover:bg-slate-50 md:flex">
                <svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M11 17h2"/><path d="M11 12h2"/><path d="M11 7h2"/></svg>
              </button>

              <div className="flex flex-row items-center gap-2">
                <button type="button" className="inline-flex h-10 items-center justify-center rounded-full bg-slate-900 px-6 text-sm font-semibold text-white shadow transition-colors hover:bg-slate-900/90">
                  Jadwalkan kunjungan
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

          <ListingDetailGallery address={address} images={images} />

          {/* Price and Specs */}
          <div className="mt-6 flex flex-col justify-between gap-6 rounded-3xl border border-border bg-(--panel) p-6 sm:flex-row sm:items-end">
            <div>
              <div className="text-4xl font-semibold tracking-tight text-foreground sm:text-5xl">{price}</div>
              <p className="mt-2 text-sm leading-7 text-muted-foreground">Ringkasan properti utama dengan fokus pada ukuran bangunan, jumlah kamar, dan kesiapan hunian.</p>
            </div>

            <div className="grid grid-cols-3 gap-6 sm:gap-10">
              <div className="flex flex-col">
                <span className="text-xl font-semibold">{beds}</span>
                <span className="text-sm font-medium uppercase tracking-widest text-muted-foreground">Kamar tidur</span>
              </div>
              <div className="flex flex-col">
                <span className="text-xl font-semibold">{baths}</span>
                <span className="text-sm font-medium uppercase tracking-widest text-muted-foreground">Kamar mandi</span>
              </div>
              <div className="flex flex-col">
                <span className="text-xl font-semibold">{buildingAreaText ?? "-"}</span>
                <span className="text-sm font-medium uppercase tracking-widest text-muted-foreground">Luas bangunan</span>
              </div>
            </div>
          </div>

          <hr className="my-12 border-border" />

          {/* About Section */}
          <section className="grid gap-8 lg:grid-cols-[280px_1fr]">
            <div>
              <h2 className="text-2xl font-light text-foreground leading-[1.1]">
                Tentang <br/><span className="text-muted-foreground">properti ini</span>
              </h2>
            </div>
            <div>
              <div className="mb-6 flex flex-wrap gap-2">
                <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-foreground shadow-sm">{category}</span>
                <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-foreground shadow-sm">{formatTransactionType(listing?.transaction_type) ?? "Jual"}</span>
                {listing?.condition ? <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-foreground shadow-sm">{formatCondition(listing.condition)}</span> : null}
                {listing?.furnishing ? <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-foreground shadow-sm">{formatFurnishing(listing.furnishing)}</span> : null}
                {listing?.year_built ? <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-foreground shadow-sm">Tahun {listing.year_built}</span> : null}
              </div>
              <p className="text-base leading-relaxed text-foreground">
                {descriptionText}
              </p>
            </div>
          </section>


          {listingVideo ? (
            <section className="mt-16 grid gap-8 lg:grid-cols-[280px_1fr]" data-testid="listing-video-tour">
              <div>
                <h2 className="text-2xl leading-[1.1] font-light text-foreground">
                  Video <br /><span className="text-muted-foreground">properti</span>
                </h2>
              </div>

              <div className="space-y-4">
                <div className="overflow-hidden rounded-3xl border border-border bg-black/90 shadow-sm">
                  <video
                    className="aspect-video h-full w-full object-cover"
                    controls
                    data-testid="listing-detail-video"
                    poster={images[0]}
                    preload="metadata"
                    src={listingVideo.url}
                  >
                    <track default kind="captions" label="Bahasa Indonesia" src="data:text/vtt,WEBVTT" srcLang="id" />
                  </video>
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-foreground shadow-sm">
                    Video penjual
                  </span>
                  {listingVideo.original_filename ? (
                    <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-muted-foreground shadow-sm">
                      {listingVideo.original_filename}
                    </span>
                  ) : null}
                  {formatVideoDuration(listingVideo.duration_seconds) ? (
                    <span className="rounded-full border border-border px-3 py-1 text-[10px] font-semibold uppercase tracking-widest text-muted-foreground shadow-sm">
                      {formatVideoDuration(listingVideo.duration_seconds)}
                    </span>
                  ) : null}
                </div>

                <p className="max-w-2xl text-sm leading-7 text-muted-foreground">
                  Lihat video walkthrough dari penjual untuk mendapatkan gambaran suasana properti sebelum menjadwalkan kunjungan langsung.
                </p>
              </div>
            </section>
          ) : null}

          <hr className="my-16 border-border" />

          {/* Other Details Section */}
          <section className="grid gap-8 lg:grid-cols-[280px_1fr]">
            <div>
              <h2 className="text-2xl font-light text-foreground leading-[1.1]">
                Detail <br/><span className="text-muted-foreground">lainnya</span>
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

          <section className="mt-16 grid gap-8 lg:grid-cols-[280px_1fr]">
            <div>
              <h2 className="text-2xl font-light leading-[1.1] text-foreground">
                Galeri <br/><span className="text-muted-foreground">properti</span>
              </h2>
            </div>
            <div className="flex items-center justify-between rounded-[1.25rem] border border-border bg-(--panel) px-5 py-4 text-sm text-muted-foreground">
              <span>{images.length} foto tersedia</span>
              <span>Gunakan panah atau klik foto utama untuk melihat detail galeri.</span>
            </div>
          </section>

          <section className="mt-16 grid gap-8 lg:grid-cols-[280px_1fr]">
            <div>
              <h2 className="text-2xl font-light leading-[1.1] text-foreground">
                Lokasi <br/><span className="text-muted-foreground">properti</span>
              </h2>
            </div>
            <div className="space-y-4">
              {locationLine ? <p className="text-sm leading-7 text-muted-foreground">{locationLine}</p> : null}
              {mapEmbedUrl ? (
                <div className="h-[300px] w-full overflow-hidden rounded-xl border border-border sm:h-[400px]">
                  <iframe
                    className="h-full w-full border-0 grayscale-10"
                    data-testid="listing-detail-map"
                    loading="lazy"
                    referrerPolicy="no-referrer-when-downgrade"
                    src={mapEmbedUrl}
                    title="Peta lokasi properti"
                  />
                </div>
              ) : (
                <div className="rounded-[1.25rem] border border-dashed border-border bg-(--panel) p-5 text-sm leading-7 text-muted-foreground">
                  Titik koordinat properti belum tersedia, jadi peta belum bisa ditampilkan secara akurat.
                </div>
              )}
            </div>
          </section>

        </div>
      </main>

      <Footer />
    </div>
  );
}
