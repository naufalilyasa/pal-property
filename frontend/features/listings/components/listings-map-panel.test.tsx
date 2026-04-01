import type { ImgHTMLAttributes } from "react";

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ListingsMapPanel } from "./listings-map-panel";

vi.mock("next/image", () => ({
  default: ({ fill: _fill, priority: _priority, unoptimized: _unoptimized, ...props }: ImgHTMLAttributes<HTMLImageElement> & { fill?: boolean; priority?: boolean; unoptimized?: boolean }) => (
    <img {...props} alt={props.alt ?? ""} />
  ),
}));

vi.mock("leaflet", () => {
  type MockMap = {
    container: HTMLElement;
    fitBounds: () => void;
    panTo: () => void;
    invalidateSize: () => void;
    remove: () => void;
  };

  const ensureMarkerPane = (container: HTMLElement) => {
    let pane = container.querySelector(".leaflet-marker-pane") as HTMLElement | null;
    if (!pane) {
      pane = document.createElement("div");
      pane.className = "leaflet-marker-pane";
      container.appendChild(pane);
    }
    return pane;
  };

  return {
    map(container: HTMLElement) {
      container.classList.add("leaflet-container");
      ensureMarkerPane(container);
      return {
        container,
        fitBounds() {},
        panTo() {},
        invalidateSize() {},
        remove() {},
      } satisfies MockMap;
    },
    tileLayer() {
      return { addTo() {} };
    },
    latLngBounds() {
      return {};
    },
    divIcon({ html }: { html: string }) {
      return { html };
    },
    marker(_latlng: [number, number], options: { icon: { html: string } }) {
      let clickHandler: (() => void) | null = null;
      let element: HTMLElement | null = null;

      const mount = (container: HTMLElement, html: string) => {
        const wrapper = document.createElement("div");
        wrapper.innerHTML = html;
        const nextElement = wrapper.firstElementChild as HTMLElement | null;
        if (!nextElement) return;
        nextElement.addEventListener("click", () => clickHandler?.());
        if (element) {
          element.replaceWith(nextElement);
        } else {
          ensureMarkerPane(container).appendChild(nextElement);
        }
        element = nextElement;
      };

      return {
        addTo(map: MockMap) {
          mount(map.container, options.icon.html);
          return this;
        },
        on(_event: string, handler: () => void) {
          clickHandler = handler;
          return this;
        },
        setIcon(icon: { html: string }) {
          const container = element?.parentElement?.parentElement?.closest(".leaflet-container") as HTMLElement | null;
          if (container) {
            mount(container, icon.html);
          }
          return this;
        },
        remove() {
          element?.remove();
          return this;
        },
      };
    },
  };
});

const listings = [
  {
    id: "listing-1",
    category_id: "cat-house",
    category: { id: "cat-house", name: "House", slug: "house" },
    title: "Jakarta River House",
    slug: "jakarta-river-house",
    transaction_type: "sale",
    price: 3250000000,
    currency: "IDR",
    location_province: "DKI Jakarta",
    location_city: "Jakarta Selatan",
    location_district: "Setiabudi",
    latitude: -6.2207,
    longitude: 106.8296,
    status: "active",
    is_featured: true,
    primary_image_url: "https://images.example/river-house.jpg",
    image_urls: ["https://images.example/river-house.jpg"],
    created_at: "2026-03-17T00:00:00Z",
    updated_at: "2026-03-17T00:00:00Z",
  },
  {
    id: "listing-2",
    category_id: "cat-house",
    category: { id: "cat-house", name: "House", slug: "house" },
    title: "Garden Court Residence",
    slug: "garden-court-residence",
    transaction_type: "sale",
    price: 2890000000,
    currency: "IDR",
    location_province: "Jawa Barat",
    location_city: "Depok",
    location_district: "Cimanggis",
    latitude: -6.3652,
    longitude: 106.9015,
    status: "active",
    is_featured: false,
    primary_image_url: "https://images.example/garden-court.jpg",
    image_urls: ["https://images.example/garden-court.jpg"],
    created_at: "2026-03-17T00:00:00Z",
    updated_at: "2026-03-17T00:00:00Z",
  },
] satisfies Parameters<typeof ListingsMapPanel>[0]["listings"];

describe("ListingsMapPanel", () => {
  it("renders clickable markers and switches popup content", async () => {
    render(<ListingsMapPanel listings={listings} />);

    await waitFor(() => {
      expect(screen.getByTestId("listing-map-marker-listing-1")).toBeInTheDocument();
      expect(screen.getByTestId("listing-map-marker-listing-2")).toBeInTheDocument();
    });
    expect(screen.getByTestId("listing-map-popup")).toHaveTextContent(/jakarta river house/i);

    fireEvent.click(screen.getByTestId("listing-map-marker-listing-2"));

    expect(screen.getByTestId("listing-map-popup")).toHaveTextContent(/garden court residence/i);
    expect(screen.getByRole("link", { name: /lihat detail/i })).toHaveAttribute("href", "/listings/garden-court-residence");
  });
});
