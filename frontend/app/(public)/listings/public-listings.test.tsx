import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import PublicListingsPage from "./page";

const {
  getOptionalUserMock,
  getSavedListingIdsForListingsMock,
  getSearchListingsMock,
} = vi.hoisted(() => ({
  getOptionalUserMock: vi.fn(),
  getSavedListingIdsForListingsMock: vi.fn(),
  getSearchListingsMock: vi.fn(),
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

vi.mock("@/features/listings/components/top-nav", () => ({
  TopNav: () => <div data-testid="top-nav" />,
}));

vi.mock("@/features/listings/components/footer", () => ({
  Footer: () => <div data-testid="footer" />,
}));

vi.mock("@/features/listings/components/listing-filters", () => ({
  ListingFilters: () => <div data-testid="listing-filters" />,
}));

vi.mock("@/features/listings/components/search-listing-card", () => ({
  SearchListingCardItem: (props: { initialSaved?: boolean; listing: { title: string } }) => (
    <article data-testid="listing-card">{`${props.listing.title}:${String(props.initialSaved ?? false)}`}</article>
  ),
}));

describe("public-listings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("prehydrates visible listing ids for signed-in users", async () => {
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
      total: 2,
      page: 1,
      limit: 12,
      total_pages: 1,
    });
    getSavedListingIdsForListingsMock.mockResolvedValue({ listingIds: ["listing-2"] });

    render(
      await PublicListingsPage({
        searchParams: Promise.resolve({}),
      }),
    );

    expect(getSavedListingIdsForListingsMock).toHaveBeenCalledWith(["listing-1", "listing-2"]);
    expect(screen.getByText("City Loft:false")).toBeInTheDocument();
    expect(screen.getByText("Garden Home:true")).toBeInTheDocument();
  });
});
