import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import PublicListingsPage from "./page";

const {
  getOptionalUserMock,
  getSavedListingIdsForListingsMock,
  getSearchListingsMock,
  listingFiltersMock,
  listingsMapPanelMock,
  searchListingCardItemMock,
} = vi.hoisted(() => ({
  getOptionalUserMock: vi.fn(),
  getSavedListingIdsForListingsMock: vi.fn(),
  getSearchListingsMock: vi.fn(),
  listingFiltersMock: vi.fn(),
  listingsMapPanelMock: vi.fn(),
  searchListingCardItemMock: vi.fn(),
}));

vi.mock("@/features/auth/server/current-user", () => ({
  getOptionalUser: getOptionalUserMock,
}));

vi.mock("@/features/saved-listings/server/get-saved-listing-ids", () => ({
  getSavedListingIdsForListings: getSavedListingIdsForListingsMock,
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
  SearchListingCardItem: (props: { href: string; initialSaved?: boolean; listing: { title: string } }) => {
    searchListingCardItemMock(props);

    return <article data-testid="listing-card">{`${props.listing.title}:${String(props.initialSaved ?? false)}`}</article>;
  },
}));

vi.mock("@/features/listings/components/listings-map-panel", () => ({
  ListingsMapPanel: (props: { listings: Array<{ title: string }> }) => {
    listingsMapPanelMock(props);

    return <div data-testid="listings-map-canvas">{`markers:${props.listings.length}`}</div>;
  },
}));

describe("PublicListingsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the listings shell with resolved query defaults", async () => {
    getOptionalUserMock.mockResolvedValue({
      id: "user-1",
      name: "Buyer",
      email: "buyer@example.com",
      role: "user",
      created_at: "2026-03-29T00:00:00Z",
    });
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
    getSavedListingIdsForListingsMock.mockResolvedValue({
      listingIds: ["listing-2"],
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
    expect(getSavedListingIdsForListingsMock).toHaveBeenCalledWith(["listing-1", "listing-2"]);
    expect(screen.getByTestId("listing-filters")).toBeInTheDocument();
    expect(screen.getByText("view:map")).toBeInTheDocument();
    expect(screen.getByTestId("listing-map-panel")).toBeInTheDocument();
    expect(screen.getByTestId("listings-map-canvas")).toHaveTextContent("markers:2");
    expect(screen.getByTestId("listing-pagination")).toBeInTheDocument();
    expect(screen.getAllByTestId("listing-card")).toHaveLength(2);
    expect(screen.getByText("City Loft:false")).toBeInTheDocument();
    expect(screen.getByText("Garden Home:true")).toBeInTheDocument();
  });

  it("renders the empty state without crashing", async () => {
    getOptionalUserMock.mockResolvedValue(null);
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
    expect(getSavedListingIdsForListingsMock).not.toHaveBeenCalled();
    expect(screen.getByText(/no properties matched your search/i)).toBeInTheDocument();
    expect(screen.getByText("view:map")).toBeInTheDocument();
  });

  it("skips saved-listing prehydration when the visitor is anonymous", async () => {
    getOptionalUserMock.mockResolvedValue(null);
    getSearchListingsMock.mockResolvedValue({
      items: [
        { id: "listing-1", slug: "city-loft", title: "City Loft", transaction_type: "sale", price: 1000, currency: "IDR", status: "active", is_featured: false, created_at: "2026-03-17T00:00:00Z", updated_at: "2026-03-17T00:00:00Z" },
      ],
      total: 1,
      page: 1,
      limit: 12,
      total_pages: 1,
    });

    render(
      await PublicListingsPage({
        searchParams: Promise.resolve({}),
      }),
    );

    expect(getSavedListingIdsForListingsMock).not.toHaveBeenCalled();
    expect(searchListingCardItemMock).toHaveBeenCalledWith(
      expect.objectContaining({
        initialSaved: false,
      }),
    );
    expect(screen.getByText("City Loft:false")).toBeInTheDocument();
  });
});
