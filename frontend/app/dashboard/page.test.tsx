import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import DashboardPage from "@/app/dashboard/page";
import { ApiError } from "@/lib/api/envelope";

const { getSellerListingsMock, getRequestCookieHeaderMock, redirectMock } = vi.hoisted(() => ({
  getSellerListingsMock: vi.fn(),
  getRequestCookieHeaderMock: vi.fn(),
  redirectMock: vi.fn((path: string) => {
    throw new Error(`NEXT_REDIRECT:${path}`);
  }),
}));

vi.mock("@/lib/api/seller-listings", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api/seller-listings")>(
    "@/lib/api/seller-listings",
  );

  return {
    ...actual,
    getSellerListings: getSellerListingsMock,
  };
});

vi.mock("@/lib/server/cookies", () => ({
  getRequestCookieHeader: getRequestCookieHeaderMock,
}));

vi.mock("next/navigation", () => ({
  redirect: redirectMock,
}));

describe("DashboardPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders seller listings for authenticated sessions", async () => {
    getRequestCookieHeaderMock.mockResolvedValue("access_token=abc");
    getSellerListingsMock.mockResolvedValue({
      data: [
        {
          id: "listing-1",
          user_id: "seller-1",
          category_id: "category-1",
          category: {
            id: "category-1",
            name: "House",
            slug: "house",
            icon_url: null,
          },
          title: "Garden Residence",
          slug: "garden-residence",
          description: null,
          price: 2750000000,
          currency: "IDR",
          location_city: "Jakarta",
          location_district: "Kebayoran",
          address_detail: null,
          status: "published",
          is_featured: false,
          specifications: {},
          view_count: 42,
          images: [
            {
              id: "image-1",
              url: "https://example.com/listing.jpg",
              format: "jpg",
              bytes: 1024,
              width: 1200,
              height: 800,
              original_filename: "listing.jpg",
              is_primary: true,
              sort_order: 1,
              created_at: "2026-03-17T00:00:00Z",
            },
          ],
          created_at: "2026-03-17T00:00:00Z",
          updated_at: "2026-03-17T00:00:00Z",
        },
      ],
      total: 1,
      page: 1,
      limit: 10,
      total_pages: 1,
    });

    render(await DashboardPage());

    expect(screen.getByRole("heading", { level: 2, name: /garden residence/i })).toBeInTheDocument();
    expect(screen.getByText(/house/i)).toBeInTheDocument();
    expect(screen.getByText(/published/i)).toBeInTheDocument();
    expect(screen.getByText(/rp\s*2.750.000.000/i)).toBeInTheDocument();
    expect(screen.getByAltText(/garden residence/i)).toBeInTheDocument();
    expect(getSellerListingsMock).toHaveBeenCalledWith({
      cookieHeader: "access_token=abc",
    });
  });

  it("redirects away when seller listings return unauthenticated", async () => {
    getRequestCookieHeaderMock.mockResolvedValue("access_token=abc");
    getSellerListingsMock.mockRejectedValue(new ApiError("unauthenticated", { status: 401 }));

    await expect(DashboardPage()).rejects.toThrow("NEXT_REDIRECT:/");
    expect(redirectMock).toHaveBeenCalledWith("/");
  });

  it("shows an empty state when the seller has no listings", async () => {
    getRequestCookieHeaderMock.mockResolvedValue("access_token=abc");
    getSellerListingsMock.mockResolvedValue({
      data: [],
      total: 0,
      page: 1,
      limit: 10,
      total_pages: 0,
    });

    render(await DashboardPage());

    expect(screen.getByRole("heading", { level: 2, name: /no listings yet/i })).toBeInTheDocument();
  });

  it("shows an error state when listings cannot be loaded", async () => {
    getRequestCookieHeaderMock.mockResolvedValue("access_token=abc");
    getSellerListingsMock.mockRejectedValue(new Error("network down"));

    render(await DashboardPage());

    expect(screen.getByRole("heading", { level: 2, name: /we could not load your listings/i })).toBeInTheDocument();
  });
});
