"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";

import type { SearchListingCard } from "@/features/listings/server/get-search-listings";

type ListingsMapPanelProps = {
  listings: SearchListingCard[];
};

type MapListing = SearchListingCard & { latitude: number; longitude: number };
type LeafletModule = typeof import("leaflet");

export function ListingsMapPanel({ listings }: ListingsMapPanelProps) {
  const markerListings = useMemo(
    () => listings.filter((listing): listing is MapListing => typeof listing.latitude === "number" && typeof listing.longitude === "number"),
    [listings],
  );
  const [activeListingId, setActiveListingId] = useState<string | null>(markerListings[0]?.id ?? null);
  const [isMapReady, setIsMapReady] = useState(false);
  const mapContainerRef = useRef<HTMLDivElement | null>(null);
  const leafletRef = useRef<LeafletModule | null>(null);
  const mapRef = useRef<import("leaflet").Map | null>(null);
  const markersRef = useRef<Map<string, import("leaflet").Marker>>(new Map());
  const activeListing = markerListings.find((listing) => listing.id === activeListingId) ?? markerListings[0] ?? null;

  useEffect(() => {
    if (!markerListings.length) {
      setActiveListingId(null);
      return;
    }

    if (!activeListingId || !markerListings.some((listing) => listing.id === activeListingId)) {
      setActiveListingId(markerListings[0].id);
    }
  }, [activeListingId, markerListings]);

  useEffect(() => {
    let cancelled = false;

    const setupMap = async () => {
      if (!mapContainerRef.current || mapRef.current) {
        return;
      }

      const leaflet = await import("leaflet");
      if (cancelled || !mapContainerRef.current) {
        return;
      }

      leafletRef.current = leaflet;

      const map = leaflet.map(mapContainerRef.current, {
        zoomControl: true,
        scrollWheelZoom: true,
      });

      leaflet
        .tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
          attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
        })
        .addTo(map);

      mapRef.current = map;
      setIsMapReady(true);
      requestAnimationFrame(() => {
        map.invalidateSize();
        requestAnimationFrame(() => map.invalidateSize());
      });
    };

    void setupMap();

    return () => {
      cancelled = true;
      markersRef.current.forEach((marker) => {
        marker.remove();
      });
      markersRef.current.clear();
      mapRef.current?.remove();
      mapRef.current = null;
      setIsMapReady(false);
    };
  }, []);

  useEffect(() => {
    const map = mapRef.current;
    const container = mapContainerRef.current;
    if (!map || !container || !isMapReady) {
      return;
    }

    const resizeObserver = new ResizeObserver(() => {
      map.invalidateSize();
    });

    resizeObserver.observe(container);
    requestAnimationFrame(() => map.invalidateSize());

    return () => {
      resizeObserver.disconnect();
    };
  }, [isMapReady]);

  useEffect(() => {
    const map = mapRef.current;
    const leaflet = leafletRef.current;
    if (!map || !leaflet || !isMapReady) {
      return;
    }

    markersRef.current.forEach((marker) => {
      marker.remove();
    });
    markersRef.current.clear();

    if (!markerListings.length) {
      return;
    }

    const bounds = leaflet.latLngBounds(markerListings.map((listing) => [listing.latitude, listing.longitude] as [number, number]));

    markerListings.forEach((listing) => {
      const marker = leaflet
        .marker([listing.latitude, listing.longitude], {
          icon: createPriceIcon(leaflet, listing, false),
          keyboard: true,
          title: listing.title,
        })
        .addTo(map)
        .on("click", () => setActiveListingId(listing.id));

      markersRef.current.set(listing.id, marker);
    });

    map.fitBounds(bounds, { padding: [24, 24], maxZoom: 12 });
  }, [isMapReady, markerListings]);

  useEffect(() => {
    const leaflet = leafletRef.current;
    const map = mapRef.current;
    if (!leaflet || !map || !activeListing || !isMapReady) {
      return;
    }

    markerListings.forEach((listing) => {
      const marker = markersRef.current.get(listing.id);
      if (!marker) {
        return;
      }

      marker.setIcon(createPriceIcon(leaflet, listing, listing.id === activeListing.id));
    });

    map.panTo([activeListing.latitude, activeListing.longitude], { animate: true, duration: 0.35 });
  }, [activeListing, isMapReady, markerListings]);

  if (markerListings.length === 0) {
    return (
      <div className="flex h-full items-center justify-center bg-[var(--panel)] p-6 text-center text-sm text-muted-foreground">
        Koordinat listing aktif belum tersedia, jadi titik properti belum bisa ditampilkan di peta.
      </div>
    );
  }

  return (
    <div className="relative h-full w-full overflow-hidden bg-[#eef2f6]">
      <div className="h-full w-full" data-testid="listings-map-canvas" ref={mapContainerRef} />

      {activeListing ? (
        <div className="pointer-events-none absolute bottom-4 left-4 right-4 z-[700]" data-testid="listing-map-popup">
          <div className="pointer-events-auto max-w-sm rounded-[1.25rem] border border-border bg-white/95 p-4 shadow-xl backdrop-blur-sm">
            <div className="flex gap-4">
              <div className="relative h-24 w-28 shrink-0 overflow-hidden rounded-xl bg-muted">
                {activeListing.primary_image_url ? (
                  <Image alt={activeListing.title} fill src={activeListing.primary_image_url} className="object-cover" unoptimized />
                ) : null}
              </div>
              <div className="min-w-0 flex-1">
                <p className="line-clamp-2 text-base font-semibold text-slate-900">{activeListing.title}</p>
                <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">
                  {[activeListing.location_district, activeListing.location_city, activeListing.location_province].filter(Boolean).join(", ")}
                </p>
                <p className="mt-3 text-sm font-semibold text-slate-900">{formatFullPrice(activeListing.price, activeListing.currency)}</p>
                <div className="mt-3 flex items-center justify-between gap-3">
                  <span className="rounded-full bg-slate-100 px-3 py-1 text-[11px] font-semibold uppercase tracking-wider text-slate-700">
                    {activeListing.category?.name ?? "Properti"}
                  </span>
                  <Link className="text-sm font-semibold text-slate-900 underline underline-offset-4" href={`/listings/${activeListing.slug}`}>
                    Lihat detail
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function createPriceIcon(leaflet: LeafletModule, listing: MapListing, selected: boolean) {
  return leaflet.divIcon({
    className: "",
    html: `<button data-testid="listing-map-marker-${listing.id}" aria-label="Lihat properti ${escapeHtml(listing.title)} di peta" class="rounded-full border-2 px-2 py-1 text-[11px] font-semibold shadow-sm transition ${selected ? "border-slate-900 bg-slate-900 text-white" : "border-white bg-white text-slate-900 hover:border-slate-300"}">${formatCompactPrice(listing.price)}</button>`,
    iconSize: [78, 34],
    iconAnchor: [39, 17],
  });
}

function formatCompactPrice(price: number) {
  if (price >= 1_000_000_000) {
    return `Rp ${(price / 1_000_000_000).toFixed(1)} M`;
  }
  if (price >= 1_000_000) {
    return `Rp ${(price / 1_000_000).toFixed(0)} Jt`;
  }
  return `Rp ${price}`;
}

function formatFullPrice(price: number, currency: string) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: currency || "IDR",
    maximumFractionDigits: 0,
  }).format(price);
}

function escapeHtml(value: string) {
  return value.replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;");
}
