import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import PublicListingsPage from "./page";

const { getListingsMock, listingFiltersMock, listingCardMock } = vi.hoisted(() => ({
  getListingsMock: vi.fn(),
  listingFiltersMock: vi.fn(),
  listingCardMock: vi.fn(),
}));

vi.mock("@/features/listings/server/get-listings", () => ({
  getListings: getListingsMock,
}));

vi.mock("@/features/listings/components/listing-filters", () => ({
  ListingFilters: (props: {
    values: {
      city?: string;
      category_id?: string;
      price_min?: string;
      price_max?: string;
      status?: string;
      limit?: string;
    };
    total: number;
    visibleCount: number;
  }) => {
    listingFiltersMock(props);

    return (
      <div data-testid="listing-filters">
        <span>{`city:${props.values.city ?? ""}`}</span>
        <span>{`status:${props.values.status ?? ""}`}</span>
        <span>{`limit:${props.values.limit ?? ""}`}</span>
        <span>{`total:${props.total}`}</span>
        <span>{`visible:${props.visibleCount}`}</span>
      </div>
    );
  },
}));

vi.mock("@/features/listings/components/listing-card", () => ({
  ListingCard: (props: { href: string; listing: { title: string } }) => {
    listingCardMock(props);
    return <article>{props.listing.title}</article>;
  },
}));

describe("PublicListingsPage", () => {
  it("renders the listings shell with resolved query defaults", async () => {
    getListingsMock.mockResolvedValue({
      data: [
        { id: "listing-1", slug: "city-loft", title: "City Loft" },
        { id: "listing-2", slug: "garden-home", title: "Garden Home" },
      ],
      total: 18,
      page: 3,
      limit: 12,
      total_pages: 2,
    });

    render(
      await PublicListingsPage({
        searchParams: Promise.resolve({
          city: "Jakarta",
          status: "active",
          limit: "12",
          category_id: "category-1",
          price_min: "500000000",
          price_max: "5000000000",
          page: "3",
        }),
      }),
    );

    expect(getListingsMock).toHaveBeenCalledWith({
      page: "3",
      limit: "12",
      city: "Jakarta",
      category_id: "category-1",
      price_min: "500000000",
      price_max: "5000000000",
      status: "active",
    });
    expect(screen.getByTestId("listing-filters")).toBeInTheDocument();
    expect(screen.getByText("city:Jakarta")).toBeInTheDocument();
    expect(screen.getByText("status:active")).toBeInTheDocument();
    expect(screen.getByText("limit:12")).toBeInTheDocument();
    expect(screen.getByTestId("listing-map-panel")).toBeInTheDocument();
    expect(screen.getByTestId("listing-pagination")).toBeInTheDocument();
    expect(screen.getAllByTestId("listing-card")).toHaveLength(2);
    expect(screen.getByText("City Loft")).toBeInTheDocument();
    expect(screen.getByText("Garden Home")).toBeInTheDocument();
  });

  it("renders the empty state without crashing", async () => {
    getListingsMock.mockResolvedValue({
      data: [],
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

    expect(getListingsMock).toHaveBeenCalledWith({
      page: "1",
      limit: "12",
      city: undefined,
      category_id: undefined,
      price_min: undefined,
      price_max: undefined,
      status: undefined,
    });
    expect(screen.getByText(/try broadening the search/i)).toBeInTheDocument();
    expect(screen.getByText("status:")).toBeInTheDocument();
    expect(screen.getByText("limit:12")).toBeInTheDocument();
    expect(screen.getByText("visible:0")).toBeInTheDocument();
  });
});
