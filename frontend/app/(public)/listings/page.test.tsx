import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import PublicListingsPage from "./page";

const { getSearchListingsMock, listingFiltersMock } = vi.hoisted(() => ({
  getSearchListingsMock: vi.fn(),
  listingFiltersMock: vi.fn(),
}));

vi.mock("@/features/listings/server/get-search-listings", () => ({
  getSearchListings: getSearchListingsMock,
}));

vi.mock("@/features/listings/components/listing-filters", () => ({
  ListingFilters: (props: { view: "map" | "list" }) => {
    listingFiltersMock(props);

    return (
      <div data-testid="listing-filters">
        <span>{`view:${props.view}`}</span>
      </div>
    );
  },
}));

vi.mock("@/features/listings/components/search-listing-card", () => ({
  SearchListingCardItem: (props: { href: string; listing: { title: string } }) => {
    return <article data-testid="listing-card">{props.listing.title}</article>;
  },
}));

describe("PublicListingsPage", () => {
  it("renders the listings shell with resolved query defaults", async () => {
    getSearchListingsMock.mockResolvedValue({
      items: [
        { id: "listing-1", slug: "city-loft", title: "City Loft", transaction_type: "sale", price: 1000, currency: "IDR", status: "active", is_featured: false, created_at: "2026-03-17T00:00:00Z", updated_at: "2026-03-17T00:00:00Z" },
        { id: "listing-2", slug: "garden-home", title: "Garden Home", transaction_type: "sale", price: 2000, currency: "IDR", status: "active", is_featured: false, created_at: "2026-03-17T00:00:00Z", updated_at: "2026-03-17T00:00:00Z" },
      ],
      total: 18,
      page: 3,
      limit: 12,
      total_pages: 2,
    });

    render(
      await PublicListingsPage({
        searchParams: Promise.resolve({
          q: "jakarta",
          transaction_type: "sale",
          limit: "12",
          category_id: "category-1",
          location_province: "DKI Jakarta",
          location_city: "Jakarta",
          price_min: "500000000",
          price_max: "5000000000",
          sort: "newest",
          page: "3",
        }),
      }),
    );

    expect(getSearchListingsMock).toHaveBeenCalledWith({
      page: "3",
      limit: "12",
      q: "jakarta",
      transaction_type: "sale",
      category_id: "category-1",
      location_province: "DKI Jakarta",
      location_city: "Jakarta",
      price_min: "500000000",
      price_max: "5000000000",
      sort: "newest",
    });
    expect(screen.getByTestId("listing-filters")).toBeInTheDocument();
    expect(screen.getByText("view:map")).toBeInTheDocument();
    expect(screen.getByTestId("listing-map-panel")).toBeInTheDocument();
    expect(screen.getByTestId("listing-pagination")).toBeInTheDocument();
    expect(screen.getAllByTestId("listing-card")).toHaveLength(2);
    expect(screen.getByText("City Loft")).toBeInTheDocument();
    expect(screen.getByText("Garden Home")).toBeInTheDocument();
  });

  it("renders the empty state without crashing", async () => {
    getSearchListingsMock.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 12,
      total_pages: 0,
    });

    render(
      await PublicListingsPage({
        searchParams: Promise.resolve({}),
      }),
    );

    expect(getSearchListingsMock).toHaveBeenCalledWith({
      page: "1",
      limit: "12",
      q: undefined,
      transaction_type: undefined,
      category_id: undefined,
      location_province: undefined,
      location_city: undefined,
      price_min: undefined,
      price_max: undefined,
      sort: undefined,
    });
    expect(screen.getByText(/no properties matched your search/i)).toBeInTheDocument();
    expect(screen.getByText("view:map")).toBeInTheDocument();
  });
});
