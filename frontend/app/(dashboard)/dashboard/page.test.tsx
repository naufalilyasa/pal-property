import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import DashboardPage from "./page";

const { getSellerListingsMock, requireUserMock } = vi.hoisted(() => ({
  getSellerListingsMock: vi.fn(),
  requireUserMock: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    refresh: vi.fn(),
  }),
}));

vi.mock("@/features/listings/server/get-seller-listings", () => ({
  getSellerListingsPage: getSellerListingsMock,
}));

vi.mock("@/features/auth/server/require-user", () => ({
  requireUser: requireUserMock,
}));

function renderWithProviders(node: React.ReactNode) {
  return render(
    <QueryClientProvider client={new QueryClient()}>{node}</QueryClientProvider>,
  );
}

describe("DashboardPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    requireUserMock.mockResolvedValue({ id: "admin-1", email: "admin@example.com" });
  });

  it("renders seller dashboard overview metrics", async () => {
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

    renderWithProviders(await DashboardPage());

    expect(screen.getByTestId("dashboard-shell")).toBeInTheDocument();
    expect(screen.getByText(/total listings/i)).toBeInTheDocument();
    expect(screen.getByText(/active page/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /review listings table/i })).toBeInTheDocument();
    expect(screen.getByTestId("dashboard-refresh-button")).toBeInTheDocument();
  });
});
