"use client";

import Image from "next/image";
import { useMemo, useState } from "react";

type ListingDetailGalleryProps = {
  address: string;
  images: string[];
};

export function ListingDetailGallery({ address, images }: ListingDetailGalleryProps) {
  const safeImages = useMemo(() => Array.from(new Set(images.filter(Boolean))), [images]);
  const [activeIndex, setActiveIndex] = useState(0);
  const [isLightboxOpen, setIsLightboxOpen] = useState(false);
  const [zoomLevel, setZoomLevel] = useState(1);

  if (safeImages.length === 0) {
    return null;
  }

  const activeImage = safeImages[activeIndex] ?? safeImages[0];
  const previewIndexes = safeImages.length > 1 ? [getWrappedIndex(activeIndex + 1, safeImages.length), getWrappedIndex(activeIndex + 2, safeImages.length)] : [];

  const move = (direction: "prev" | "next") => {
    setActiveIndex((current) => getWrappedIndex(current + (direction === "next" ? 1 : -1), safeImages.length));
    setZoomLevel(1);
  };

  const adjustZoom = (nextZoom: number) => {
    setZoomLevel(Math.max(1, Math.min(nextZoom, 3)));
  };

  return (
    <>
      <section className="mt-6 grid gap-4 lg:grid-cols-[72px_minmax(0,1fr)_320px]" data-testid="listing-detail-gallery">
        <div className="hidden lg:flex lg:flex-col lg:gap-3">
          {safeImages.slice(0, 6).map((image, index) => {
            const selected = index === activeIndex;

            return (
              <button
                key={`thumb-${image}`}
                aria-label={`Pilih foto ${index + 1}`}
                className={`relative h-20 overflow-hidden rounded-[1rem] border transition ${selected ? "border-slate-900 shadow-sm" : "border-border hover:border-slate-400"}`}
                onClick={() => setActiveIndex(index)}
                type="button"
              >
                <Image alt={`${address} thumbnail ${index + 1}`} fill src={image} className="object-cover" unoptimized />
              </button>
            );
          })}
        </div>

        <div className="relative overflow-hidden rounded-[1.5rem] border border-border bg-muted">
          <button
            aria-label="Buka galeri detail"
            className="relative block aspect-[16/10] w-full"
            data-testid="listing-detail-gallery-open"
            onClick={() => {
              setZoomLevel(1);
              setIsLightboxOpen(true);
            }}
            type="button"
          >
            <Image alt={address} fill priority src={activeImage} className="object-cover" unoptimized />
          </button>

          {safeImages.length > 1 ? (
            <>
              <button
                aria-label="Foto sebelumnya"
                className="absolute left-4 top-1/2 flex h-11 w-11 -translate-y-1/2 items-center justify-center rounded-full bg-white/90 text-slate-900 shadow-sm transition hover:bg-white"
                data-testid="listing-detail-gallery-prev"
                onClick={() => move("prev")}
                type="button"
              >
                ←
              </button>
              <button
                aria-label="Foto berikutnya"
                className="absolute right-4 top-1/2 flex h-11 w-11 -translate-y-1/2 items-center justify-center rounded-full bg-white/90 text-slate-900 shadow-sm transition hover:bg-white"
                data-testid="listing-detail-gallery-next"
                onClick={() => move("next")}
                type="button"
              >
                →
              </button>
            </>
          ) : null}

          <div className="absolute inset-x-0 bottom-0 flex items-center justify-between bg-gradient-to-t from-black/65 via-black/10 to-transparent px-4 py-4 text-sm text-white">
            <span>
              {String(activeIndex + 1).padStart(2, "0")} / {String(safeImages.length).padStart(2, "0")}
            </span>
            <span>Klik foto untuk melihat detail</span>
          </div>
        </div>

        <div className="hidden gap-4 lg:grid">
          {previewIndexes.map((index) => (
            <button
              key={`${safeImages[index]}-${index}`}
              aria-label={`Lihat foto ${index + 1}`}
              className="relative aspect-[4/3] overflow-hidden rounded-[1.5rem] border border-border bg-muted"
              onClick={() => setActiveIndex(index)}
              type="button"
            >
              <Image alt={`${address} preview ${index + 1}`} fill src={safeImages[index]} className="object-cover" unoptimized />
            </button>
          ))}
        </div>
      </section>

      {isLightboxOpen ? (
        <div
          aria-label="Galeri foto properti"
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/85 px-4 py-6"
          data-testid="listing-detail-gallery-lightbox"
          role="dialog"
        >
          <button
            aria-label="Tutup galeri"
            className="absolute right-6 top-6 flex h-12 w-12 items-center justify-center rounded-full bg-white text-slate-900 shadow-sm"
            onClick={() => {
              setZoomLevel(1);
              setIsLightboxOpen(false);
            }}
            type="button"
          >
            ✕
          </button>

          <div className="absolute left-6 top-6 z-10 flex items-center gap-2 rounded-full bg-white/10 px-3 py-2 text-white backdrop-blur-sm">
            <button
              aria-label="Zoom out foto"
              className="flex h-10 w-10 items-center justify-center rounded-full bg-white text-slate-900 shadow-sm"
              onClick={() => adjustZoom(zoomLevel - 0.25)}
              type="button"
            >
              −
            </button>
            <button
              aria-label="Reset zoom foto"
              className="rounded-full bg-white/10 px-3 py-2 text-sm font-semibold text-white"
              onClick={() => setZoomLevel(1)}
              type="button"
            >
              {Math.round(zoomLevel * 100)}%
            </button>
            <button
              aria-label="Zoom in foto"
              className="flex h-10 w-10 items-center justify-center rounded-full bg-white text-slate-900 shadow-sm"
              onClick={() => adjustZoom(zoomLevel + 0.25)}
              type="button"
            >
              +
            </button>
          </div>

          <div className="grid w-full max-w-[1500px] gap-4 lg:grid-cols-[96px_minmax(0,1fr)_220px]">
            <div className="hidden max-h-[80vh] gap-3 overflow-y-auto lg:grid">
              {safeImages.map((image, index) => {
                const selected = index === activeIndex;

                return (
                  <button
                    key={`lightbox-${image}`}
                    aria-label={`Pilih foto ${index + 1} di galeri`}
                    className={`relative h-24 overflow-hidden rounded-[1rem] border transition ${selected ? "border-white" : "border-white/20 hover:border-white/60"}`}
                    onClick={() => {
                      setActiveIndex(index);
                      setZoomLevel(1);
                    }}
                    type="button"
                  >
                    <Image alt={`${address} lightbox thumbnail ${index + 1}`} fill src={image} className="object-cover" unoptimized />
                  </button>
                );
              })}
            </div>

            <div className="relative overflow-hidden rounded-[1.5rem] bg-black">
              <div className="relative aspect-[16/10] w-full max-h-[80vh] min-h-[320px] overflow-auto">
                <div
                  className="relative h-full w-full transition-transform duration-200"
                  style={{ transform: `scale(${zoomLevel})`, transformOrigin: "center center" }}
                >
                  <Image alt={`${address} detail ${activeIndex + 1}`} fill src={activeImage} className="object-contain" unoptimized />
                </div>
              </div>

              {safeImages.length > 1 ? (
                <>
                  <button
                    aria-label="Foto sebelumnya di galeri"
                    className="absolute left-4 top-1/2 flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full bg-white text-slate-900 shadow-sm"
                    onClick={() => move("prev")}
                    type="button"
                  >
                    ←
                  </button>
                  <button
                    aria-label="Foto berikutnya di galeri"
                    className="absolute right-4 top-1/2 flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full bg-white text-slate-900 shadow-sm"
                    onClick={() => move("next")}
                    type="button"
                  >
                    →
                  </button>
                </>
              ) : null}
            </div>

            <div className="hidden lg:block">
              <div className="rounded-[1.5rem] bg-white/10 p-5 text-white backdrop-blur-sm">
                <p className="text-3xl font-semibold leading-none">{String(activeIndex + 1).padStart(2, "0")}</p>
                <p className="mt-2 text-base text-white/70">/ {String(safeImages.length).padStart(2, "0")}</p>
                <div className="mt-8 h-1 rounded-full bg-white/15">
                  <div
                    className="h-full rounded-full bg-white transition-[width]"
                    style={{ width: `${((activeIndex + 1) / safeImages.length) * 100}%` }}
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}

function getWrappedIndex(index: number, length: number) {
  return (index + length) % length;
}
