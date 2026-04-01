import type { ImgHTMLAttributes } from "react";

import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import PublicListingDetailPage from "./page";

const {
  getListingBySlugMock,
  getOptionalUserMock,
  getSavedListingIdsForListingsMock,
  saveListingButtonMock,
} = vi.hoisted(() => ({
  getListingBySlugMock: vi.fn(),
  getOptionalUserMock: vi.fn(),
  getSavedListingIdsForListingsMock: vi.fn(),
  saveListingButtonMock: vi.fn(),
}));

vi.mock("next/image", () => ({
  default: ({
    fill: _fill,
    priority: _priority,
    unoptimized: _unoptimized,
    ...props
  }: ImgHTMLAttributes<HTMLImageElement> & {
    fill?: boolean;
    priority?: boolean;
    unoptimized?: boolean;
  }) => <img {...props} alt={props.alt ?? ""} />,
}));

vi.mock("@/features/auth/server/current-user", () => ({
  getOptionalUser: getOptionalUserMock,
}));

vi.mock("@/features/listings/server/get-listing-by-slug", () => ({
  getListingBySlug: getListingBySlugMock,
}));

vi.mock("@/features/saved-listings/server/get-saved-listing-ids", () => ({
  getSavedListingIdsForListings: getSavedListingIdsForListingsMock,
}));

vi.mock("@/features/saved-listings/components/save-listing-button", () => ({
  SaveListingButton: (props: { initialSaved: boolean; listingId: string; variant: string }) => {
    saveListingButtonMock(props);

    return (
      <button data-pressed={String(props.initialSaved)} data-testid="save-listing-button" type="button">
        {`${props.listingId}:${String(props.initialSaved)}:${props.variant}`}
      </button>
    );
  },
}));

describe("PublicListingDetailPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("prehydrates the current listing save state for authenticated users", async () => {
    getOptionalUserMock.mockResolvedValue({
      id: "user-1",
      name: "Buyer",
      email: "buyer@example.com",
      role: "user",
      created_at: "2026-03-29T00:00:00Z",
    });
    getListingBySlugMock.mockResolvedValue({
      id: "listing-1",
      user_id: "seller-1",
      title: "Jakarta River House",
      slug: "jakarta-river-house",
      price: 3250000000,
      currency: "IDR",
      status: "active",
      is_featured: true,
      latitude: -6.2088,
      longitude: 106.8456,
      specifications: {},
      view_count: 0,
      images: [],
      created_at: "2026-03-17T00:00:00Z",
      updated_at: "2026-03-17T00:00:00Z",
    });
    getSavedListingIdsForListingsMock.mockResolvedValue({ listingIds: ["listing-1"] });

    render(
      await PublicListingDetailPage({
        params: Promise.resolve({ slug: "jakarta-river-house" }),
      }),
    );

    expect(getListingBySlugMock).toHaveBeenCalledWith("jakarta-river-house");
    expect(getSavedListingIdsForListingsMock).toHaveBeenCalledWith(["listing-1"]);
    expect(saveListingButtonMock).toHaveBeenCalledWith(
      expect.objectContaining({
        initialSaved: true,
        listingId: "listing-1",
        variant: "icon",
      }),
    );
    expect(screen.getByTestId("save-listing-button")).toHaveTextContent("listing-1:true:icon");
  });

  it("renders the reusable save control for anonymous visitors without prehydration fetches", async () => {
    getOptionalUserMock.mockResolvedValue(null);
    getListingBySlugMock.mockResolvedValue({
      id: "listing-2",
      user_id: "seller-1",
      title: "Garden Court Residence",
      slug: "garden-court-residence",
      price: 2890000000,
      currency: "IDR",
      status: "active",
      is_featured: false,
      specifications: {},
      view_count: 0,
      images: [],
      video: null,
      created_at: "2026-03-17T00:00:00Z",
      updated_at: "2026-03-17T00:00:00Z",
    });

    render(
      await PublicListingDetailPage({
        params: Promise.resolve({ slug: "garden-court-residence" }),
      }),
    );

    expect(getSavedListingIdsForListingsMock).not.toHaveBeenCalled();
    expect(saveListingButtonMock).toHaveBeenCalledWith(
      expect.objectContaining({
        initialSaved: false,
        listingId: "listing-2",
        variant: "icon",
      }),
    );
    expect(screen.getByTestId("save-listing-button")).toHaveTextContent("listing-2:false:icon");
    expect(screen.queryByRole("heading", { level: 2, name: /video tour/i })).not.toBeInTheDocument();
  });

  it("renders a video tour block only when listing.video exists", async () => {
    getOptionalUserMock.mockResolvedValue(null);
    getListingBySlugMock.mockResolvedValue({
      id: "listing-3",
      user_id: "seller-1",
      title: "Clifftop Escape",
      slug: "clifftop-escape",
      price: 4100000000,
      currency: "IDR",
      status: "active",
      is_featured: true,
      specifications: {},
      view_count: 0,
      images: [
        {
          id: "image-1",
          url: "https://images.example/clifftop.jpg",
          is_primary: true,
          sort_order: 0,
          created_at: "2026-03-17T00:00:00Z",
        },
      ],
      video: {
        id: "video-1",
        url: "https://videos.example/clifftop-tour.mp4",
        original_filename: "clifftop-tour.mp4",
        duration_seconds: 58,
        created_at: "2026-03-17T00:00:01Z",
      },
      created_at: "2026-03-17T00:00:00Z",
      updated_at: "2026-03-17T00:00:00Z",
    });

    render(
      await PublicListingDetailPage({
        params: Promise.resolve({ slug: "clifftop-escape" }),
      }),
    );

    expect(screen.getByRole("heading", { level: 2, name: /video properti/i })).toBeInTheDocument();
    expect(screen.getByTestId("listing-video-tour")).toBeInTheDocument();
    expect(screen.getByTestId("listing-detail-video")).toHaveAttribute("src", "https://videos.example/clifftop-tour.mp4");
    expect(screen.getByText(/clifftop-tour\.mp4/i)).toBeInTheDocument();
  });

  it("shows land area derived from legacy specifications when direct field is absent", async () => {
    getOptionalUserMock.mockResolvedValue(null);
    getListingBySlugMock.mockResolvedValue({
      id: "listing-4",
      user_id: "seller-1",
      title: "Serenity Villa",
      slug: "serenity-villa",
      price: 5200000000,
      currency: "IDR",
      status: "active",
      is_featured: false,
      latitude: -6.1754,
      longitude: 106.8272,
      specifications: { land_area_sqm: 180 },
      view_count: 0,
      images: [],
      created_at: "2026-03-17T00:00:00Z",
      updated_at: "2026-03-17T00:00:00Z",
    });

    render(
      await PublicListingDetailPage({
        params: Promise.resolve({ slug: "serenity-villa" }),
      }),
    );

    expect(screen.getByText(/luas tanah/i)).toBeInTheDocument();
    expect(screen.getByText("180 m²")).toBeInTheDocument();
  });

  it("renders the map using the listing coordinates", async () => {
    getOptionalUserMock.mockResolvedValue(null);
    getListingBySlugMock.mockResolvedValue({
      id: "listing-5",
      user_id: "seller-1",
      title: "Menteng Residence",
      slug: "menteng-residence",
      price: 7800000000,
      currency: "IDR",
      status: "active",
      is_featured: true,
      latitude: -6.1961,
      longitude: 106.8296,
      specifications: {},
      view_count: 0,
      images: [],
      created_at: "2026-03-17T00:00:00Z",
      updated_at: "2026-03-17T00:00:00Z",
    });

    render(
      await PublicListingDetailPage({
        params: Promise.resolve({ slug: "menteng-residence" }),
      }),
    );

    expect(screen.getByTestId("listing-detail-map")).toHaveAttribute(
      "src",
      expect.stringContaining("marker=-6.196100%2C106.829600"),
    );
  });
});
