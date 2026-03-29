import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import SavedListingsPage from "./page";

const { getSavedListingsPageMock, requireUserMock, searchListingCardItemMock } = vi.hoisted(() => ({
  getSavedListingsPageMock: vi.fn(),
  requireUserMock: vi.fn(),
  searchListingCardItemMock: vi.fn(),
}));

vi.mock("@/features/auth/server/require-user", () => ({
  requireUser: requireUserMock,
}));

vi.mock("@/features/saved-listings/server/get-saved-listings", () => ({
  getSavedListingsPage: getSavedListingsPageMock,
}));

vi.mock("@/features/listings/components/top-nav", () => ({
  TopNav: () => <div data-testid="top-nav" />,
}));

vi.mock("@/features/listings/components/footer", () => ({
  Footer: () => <div data-testid="footer" />,
}));

vi.mock("@/features/listings/components/search-listing-card", () => ({
  SearchListingCardItem: (props: { href: string; initialSaved?: boolean; refreshOnRemove?: boolean; listing: { title: string } }) => {
    searchListingCardItemMock(props);

    return <article data-testid="listing-card">{props.listing.title}</article>;
  },
}));

describe("SavedListingsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    requireUserMock.mockResolvedValue({
      id: "user-1",
      email: "buyer@example.com",
    });
  });

  it("renders the protected saved listings grid with default page and limit params", async () => {
    getSavedListingsPageMock.mockResolvedValue({
      data: [
        buildSavedListing({
          id: "listing-1",
          slug: "newest-loft",
          title: "Newest Loft",
          images: [
            {
              id: "image-1",
              url: "https://images.example/newest-primary.jpg",
              is_primary: true,
              sort_order: 0,
              created_at: "2026-03-17T00:00:00Z",
            },
          ],
        }),
        buildSavedListing({
          id: "listing-2",
          slug: "garden-retreat",
          title: "Garden Retreat",
          category: null,
          images: [
            {
              id: "image-2",
              url: "https://images.example/garden-secondary.jpg",
              is_primary: false,
              sort_order: 0,
              created_at: "2026-03-17T00:00:00Z",
            },
          ],
        }),
      ],
      total: 2,
      page: 1,
      limit: 12,
      total_pages: 2,
    });

    render(await SavedListingsPage({ searchParams: Promise.resolve({}) }));

    expect(requireUserMock).toHaveBeenCalledTimes(1);
    expect(getSavedListingsPageMock).toHaveBeenCalledWith({ page: "1", limit: "12" });
    expect(screen.getByRole("heading", { level: 1, name: /saved listings, ready when you are/i })).toBeInTheDocument();
    expect(screen.getByText(/newest saved first/i)).toBeInTheDocument();
    expect(screen.getByTestId("saved-listings-grid")).toBeInTheDocument();
    expect(screen.getByTestId("saved-listings-pagination")).toBeInTheDocument();
    expect(screen.getAllByTestId("listing-card")).toHaveLength(2);
    expect(searchListingCardItemMock).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({
        href: "/listings/newest-loft",
        initialSaved: true,
        refreshOnRemove: true,
        listing: expect.objectContaining({
          id: "listing-1",
          primary_image_url: "https://images.example/newest-primary.jpg",
          image_urls: ["https://images.example/newest-primary.jpg"],
          category: expect.objectContaining({ name: "House" }),
        }),
      }),
    );
    expect(searchListingCardItemMock).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({
        href: "/listings/garden-retreat",
        listing: expect.objectContaining({
          id: "listing-2",
          primary_image_url: "https://images.example/garden-secondary.jpg",
          category: undefined,
        }),
      }),
    );
  });

  it("renders the empty state with a browse CTA when there are no saved listings", async () => {
    getSavedListingsPageMock.mockResolvedValue({
      data: [],
      total: 0,
      page: 1,
      limit: 12,
      total_pages: 0,
    });

    render(
      await SavedListingsPage({
        searchParams: Promise.resolve({ page: "3", limit: "12" }),
      }),
    );

    expect(getSavedListingsPageMock).toHaveBeenCalledWith({ page: "3", limit: "12" });
    expect(screen.getByTestId("saved-listings-empty")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /explore listings/i })).toHaveAttribute("href", "/listings");
    expect(screen.queryByTestId("saved-listings-grid")).not.toBeInTheDocument();
  });

  it("bubbles the requireUser redirect when the visitor is unauthenticated", async () => {
    const redirectError = new Error("NEXT_REDIRECT");
    requireUserMock.mockRejectedValueOnce(redirectError);

    await expect(SavedListingsPage({ searchParams: Promise.resolve({}) })).rejects.toBe(redirectError);
    expect(getSavedListingsPageMock).not.toHaveBeenCalled();
  });
});

function buildSavedListing(
  overrides: Partial<{
    id: string;
    slug: string;
    title: string;
    description: string | null;
    category_id: string | null;
    category: { id: string; name: string; slug: string; icon_url?: string | null } | null;
    images: Array<{
      id: string;
      url: string;
      is_primary: boolean;
      sort_order: number;
      created_at: string;
      format?: string | null;
      bytes?: number | null;
      width?: number | null;
      height?: number | null;
      original_filename?: string | null;
    }>;
  }> = {},
) {
  return {
    id: "listing-1",
    user_id: "user-1",
    category_id: "category-1",
    category: { id: "category-1", name: "House", slug: "house", icon_url: null },
    title: "Saved Listing",
    slug: "saved-listing",
    description: "Quiet street, quick commute.",
    price: 2750000000,
    currency: "IDR",
    location_city: "Jakarta",
    location_district: "Menteng",
    address_detail: "Jl. Example 12",
    status: "active",
    is_featured: false,
    specifications: {},
    view_count: 120,
    images: [],
    created_at: "2026-03-17T00:00:00Z",
    updated_at: "2026-03-17T00:00:00Z",
    ...overrides,
  };
}
