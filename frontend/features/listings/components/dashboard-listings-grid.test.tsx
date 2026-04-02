import type { ImgHTMLAttributes } from "react";

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { DashboardListingsGrid } from "./dashboard-listings-grid";

const { deleteSellerListingMock, refreshMock, updateSellerListingStatusMock } = vi.hoisted(() => ({
  deleteSellerListingMock: vi.fn(),
  refreshMock: vi.fn(),
  updateSellerListingStatusMock: vi.fn(),
}));

vi.mock("next/image", () => ({
  default: ({ fill: _fill, unoptimized: _unoptimized, ...props }: ImgHTMLAttributes<HTMLImageElement> & { fill?: boolean; unoptimized?: boolean }) => <img {...props} alt={props.alt ?? ""} />,
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh: refreshMock }),
}));

vi.mock("@/lib/api/seller-listings-client", () => ({
  deleteSellerListing: deleteSellerListingMock,
  updateSellerListingStatus: updateSellerListingStatusMock,
}));

const listings = [
  {
    id: "listing-1",
    user_id: "seller-1",
    category_id: "cat-house",
    category: { id: "cat-house", name: "House", slug: "house" },
    title: "Garden Residence",
    slug: "garden-residence",
    description: "Family house",
    price: 2750000000,
    currency: "IDR",
    location_city: "Jakarta",
    location_district: "Kebayoran",
    address_detail: null,
    status: "active",
    is_featured: false,
    specifications: {},
    view_count: 7,
    images: [{ id: "image-1", url: "https://images.example/1.jpg", is_primary: true, sort_order: 0, created_at: "2026-03-17T00:00:00Z" }],
    video: null,
    created_at: "2026-03-17T00:00:00Z",
    updated_at: "2026-03-17T00:00:00Z",
  },
] as const;

describe("DashboardListingsGrid", () => {
  beforeEach(() => {
    deleteSellerListingMock.mockReset();
    refreshMock.mockReset();
    updateSellerListingStatusMock.mockReset();
    vi.spyOn(window, "confirm").mockReturnValue(true);
  });

  it("renders seller listings as cards", () => {
    render(<DashboardListingsGrid listings={[...listings]} />);

    expect(screen.getByTestId("dashboard-listings-grid")).toBeInTheDocument();
    expect(screen.getByTestId("dashboard-listing-card-listing-1")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /mark sold/i })).toBeInTheDocument();
  });

  it("marks a listing as sold", async () => {
    updateSellerListingStatusMock.mockResolvedValue({ ...listings[0], status: "sold" });

    render(<DashboardListingsGrid listings={[...listings]} />);
    fireEvent.click(screen.getByRole("button", { name: /mark sold/i }));

    await waitFor(() => {
      expect(updateSellerListingStatusMock).toHaveBeenCalledWith("listing-1", "sold");
      expect(refreshMock).toHaveBeenCalled();
      expect(screen.getByText(/terjual/i)).toBeInTheDocument();
    });
  });

  it("archives and deletes a listing", async () => {
    updateSellerListingStatusMock.mockResolvedValue({ ...listings[0], status: "archived" });
    deleteSellerListingMock.mockResolvedValue(undefined);

    render(<DashboardListingsGrid listings={[...listings]} />);
    fireEvent.click(screen.getByRole("button", { name: /archive/i }));

    await waitFor(() => expect(updateSellerListingStatusMock).toHaveBeenCalledWith("listing-1", "archived"));

    fireEvent.click(screen.getByRole("button", { name: /delete/i }));
    await waitFor(() => {
      expect(deleteSellerListingMock).toHaveBeenCalledWith("listing-1");
      expect(screen.queryByTestId("dashboard-listing-card-listing-1")).not.toBeInTheDocument();
    });
  });
});
