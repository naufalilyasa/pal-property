import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/envelope";

import { ListingForm } from "./listing-form";

const { pushMock, refreshMock } = vi.hoisted(() => ({
  pushMock: vi.fn(),
  refreshMock: vi.fn(),
}));

const {
  createSellerListingMock,
  deleteListingImageMock,
  getListingByIdMock,
  getListingCategoriesMock,
  reorderListingImagesMock,
  setPrimaryListingImageMock,
  uploadListingImageMock,
  updateSellerListingMock,
} = vi.hoisted(() => ({
  createSellerListingMock: vi.fn(),
  deleteListingImageMock: vi.fn(),
  getListingByIdMock: vi.fn(),
  getListingCategoriesMock: vi.fn(),
  reorderListingImagesMock: vi.fn(),
  setPrimaryListingImageMock: vi.fn(),
  uploadListingImageMock: vi.fn(),
  updateSellerListingMock: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: pushMock,
    refresh: refreshMock,
  }),
}));

vi.mock("@/lib/api/listing-form", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api/listing-form")>(
    "@/lib/api/listing-form",
  );

  return {
    ...actual,
    createSellerListing: createSellerListingMock,
    deleteListingImage: deleteListingImageMock,
    getListingById: getListingByIdMock,
    getListingCategories: getListingCategoriesMock,
    reorderListingImages: reorderListingImagesMock,
    setPrimaryListingImage: setPrimaryListingImageMock,
    uploadListingImage: uploadListingImageMock,
    updateSellerListing: updateSellerListingMock,
  };
});

function buildListing(overrides: Partial<import("@/lib/api/listing-form").ListingRecord> = {}) {
  return {
    id: "listing-7",
    user_id: "seller-1",
    category_id: "cat-child",
    category: { id: "cat-child", name: "Villa", slug: "villa", icon_url: null },
    title: "Existing Residence",
    slug: "existing-residence",
    description: "Fresh paint and pool.",
    price: 3150000000,
    currency: "IDR",
    location_city: "Bandung",
    location_district: "Cidadap",
    address_detail: "Jl. Setiabudi 10",
    status: "inactive",
    is_featured: false,
    specifications: {
      bedrooms: 5,
      bathrooms: 3,
      land_area_sqm: 240,
      building_area_sqm: 180,
    },
    view_count: 12,
    images: [],
    created_at: "2026-03-17T00:00:00Z",
    updated_at: "2026-03-17T00:00:00Z",
    ...overrides,
  };
}

function renderWithProviders(node: React.ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
        retry: false,
      },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>{node}</QueryClientProvider>,
  );
}

describe("ListingForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    getListingCategoriesMock.mockResolvedValue([
      { id: "cat-root", name: "House", slug: "house", label: "House" },
      { id: "cat-child", name: "Villa", slug: "villa", label: "House / Villa" },
    ]);
  });

  it("loads create mode categories and submits the canonical create payload", async () => {
    createSellerListingMock.mockResolvedValue({ id: "listing-99" });

    renderWithProviders(<ListingForm mode="create" />);

    expect(await screen.findByRole("heading", { level: 2, name: /publish a new property draft/i })).toBeInTheDocument();
    expect(await screen.findByRole("option", { name: /house \/ villa/i })).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/^title/i), { target: { value: "Garden Residence" } });
    fireEvent.change(screen.getByLabelText(/^price/i), { target: { value: "2750000000" } });
    fireEvent.change(screen.getByLabelText(/^category/i), { target: { value: "cat-child" } });
    fireEvent.change(screen.getByLabelText(/^city/i), { target: { value: "Jakarta" } });
    fireEvent.change(screen.getByLabelText(/^bedrooms/i), { target: { value: "4" } });

    fireEvent.click(screen.getByRole("button", { name: /create listing/i }));

    await waitFor(() => {
      expect(createSellerListingMock).toHaveBeenCalledWith({
        category_id: "cat-child",
        title: "Garden Residence",
        description: null,
        price: 2750000000,
        location_city: "Jakarta",
        location_district: null,
        address_detail: null,
        status: "active",
        specifications: {
          bedrooms: 4,
          bathrooms: 0,
          land_area_sqm: 0,
          building_area_sqm: 0,
        },
      });
    });
    expect(pushMock).toHaveBeenCalledWith("/dashboard/listings/listing-99/edit?created=1");
  });

  it("normalizes create payload values before submitting to backend", async () => {
    createSellerListingMock.mockResolvedValue({ id: "listing-normalize" });

    renderWithProviders(<ListingForm mode="create" />);

    await screen.findByRole("heading", { level: 2, name: /publish a new property draft/i });

    fireEvent.change(screen.getByLabelText(/^title/i), { target: { value: "  Sunset Loft  " } });
    fireEvent.change(screen.getByLabelText(/^description/i), {
      target: { value: "   " },
    });
    fireEvent.change(screen.getByLabelText(/^price/i), { target: { value: "0" } });
    fireEvent.change(screen.getByLabelText(/^city/i), { target: { value: "  Surabaya " } });
    fireEvent.change(screen.getByLabelText(/^district/i), { target: { value: "   " } });
    fireEvent.change(screen.getByLabelText(/^address detail/i), { target: { value: "  Tower A  " } });
    fireEvent.change(screen.getByLabelText(/^status/i), { target: { value: "sold" } });
    fireEvent.change(screen.getByLabelText(/^bedrooms/i), { target: { value: "" } });
    fireEvent.change(screen.getByLabelText(/^bathrooms/i), { target: { value: "" } });
    fireEvent.change(screen.getByLabelText(/^land area/i), { target: { value: "120" } });

    fireEvent.click(screen.getByRole("button", { name: /create listing/i }));

    await waitFor(() => {
      expect(createSellerListingMock).toHaveBeenCalledWith({
        category_id: null,
        title: "Sunset Loft",
        description: null,
        price: 1,
        location_city: "Surabaya",
        location_district: null,
        address_detail: "Tower A",
        status: "sold",
        specifications: {
          bedrooms: 0,
          bathrooms: 0,
          land_area_sqm: 120,
          building_area_sqm: 0,
        },
      });
    });
  });

  it("hydrates edit mode from seller-owned listing data and saves updates", async () => {
    const initialListing = buildListing();
    updateSellerListingMock.mockResolvedValue(
      buildListing({
        category_id: "cat-root",
        category: { id: "cat-root", name: "House", slug: "house", icon_url: null },
        title: "Existing Residence Updated",
        price: 3300000000,
      }),
    );

    renderWithProviders(<ListingForm initialListing={initialListing} listingId="listing-7" mode="edit" />);

    expect(await screen.findByDisplayValue("Existing Residence")).toBeInTheDocument();
    expect(await screen.findByRole("option", { name: /^house$/i })).toBeInTheDocument();
    expect(screen.getByDisplayValue("Bandung")).toBeInTheDocument();
    expect(screen.getByDisplayValue("3150000000")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/^title/i), { target: { value: "Existing Residence Updated" } });
    fireEvent.change(screen.getByLabelText(/^price/i), { target: { value: "3300000000" } });
    fireEvent.change(screen.getByLabelText(/^category/i), { target: { value: "cat-root" } });

    fireEvent.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      expect(updateSellerListingMock).toHaveBeenCalledWith("listing-7", {
        category_id: "cat-root",
        title: "Existing Residence Updated",
        description: "Fresh paint and pool.",
        price: 3300000000,
        location_city: "Bandung",
        location_district: "Cidadap",
        address_detail: "Jl. Setiabudi 10",
        status: "inactive",
        specifications: {
          bedrooms: 5,
          bathrooms: 3,
          land_area_sqm: 240,
          building_area_sqm: 180,
        },
      });
    });

    expect(await screen.findByText(/listing changes saved successfully/i)).toBeInTheDocument();
    expect(refreshMock).toHaveBeenCalled();
    expect(getListingByIdMock).not.toHaveBeenCalled();
  });

  it("surfaces backend validation and transport errors clearly", async () => {
    createSellerListingMock.mockRejectedValue(
      new ApiError("title must be at least 5 characters", {
        status: 400,
        traceId: "trace-422",
      }),
    );

    renderWithProviders(<ListingForm mode="create" />);

    await screen.findByRole("heading", { level: 2, name: /publish a new property draft/i });

    fireEvent.change(screen.getByLabelText(/^title/i), { target: { value: "Tiny home" } });
    fireEvent.change(screen.getByLabelText(/^price/i), { target: { value: "950000000" } });

    fireEvent.click(screen.getByRole("button", { name: /create listing/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      /title must be at least 5 characters \(trace trace-422\)/i,
    );
  });

  it("uploads, reorders, sets primary, and deletes images from backend responses", async () => {
    const initialListing = buildListing({
      images: [
        {
          id: "image-1",
          url: "https://images.example/1.jpg",
          original_filename: "front.jpg",
          is_primary: true,
          sort_order: 0,
          created_at: "2026-03-17T00:00:00Z",
        },
        {
          id: "image-2",
          url: "https://images.example/2.jpg",
          original_filename: "pool.jpg",
          is_primary: false,
          sort_order: 1,
          created_at: "2026-03-17T00:00:01Z",
        },
      ],
    });
    uploadListingImageMock.mockResolvedValue(
      buildListing({
        images: [
          {
            id: "image-1",
            url: "https://images.example/1.jpg",
            original_filename: "front.jpg",
            is_primary: true,
            sort_order: 0,
            created_at: "2026-03-17T00:00:00Z",
          },
          {
            id: "image-2",
            url: "https://images.example/2.jpg",
            original_filename: "pool.jpg",
            is_primary: false,
            sort_order: 1,
            created_at: "2026-03-17T00:00:01Z",
          },
          {
            id: "image-3",
            url: "https://images.example/3.jpg",
            original_filename: "garden.png",
            is_primary: false,
            sort_order: 2,
            created_at: "2026-03-17T00:00:02Z",
          },
        ],
      }),
    );
    reorderListingImagesMock.mockResolvedValue(
      buildListing({
        images: [
          {
            id: "image-2",
            url: "https://images.example/2.jpg",
            original_filename: "pool.jpg",
            is_primary: false,
            sort_order: 0,
            created_at: "2026-03-17T00:00:01Z",
          },
          {
            id: "image-1",
            url: "https://images.example/1.jpg",
            original_filename: "front.jpg",
            is_primary: true,
            sort_order: 1,
            created_at: "2026-03-17T00:00:00Z",
          },
          {
            id: "image-3",
            url: "https://images.example/3.jpg",
            original_filename: "garden.png",
            is_primary: false,
            sort_order: 2,
            created_at: "2026-03-17T00:00:02Z",
          },
        ],
      }),
    );
    setPrimaryListingImageMock.mockResolvedValue(
      buildListing({
        images: [
          {
            id: "image-2",
            url: "https://images.example/2.jpg",
            original_filename: "pool.jpg",
            is_primary: true,
            sort_order: 0,
            created_at: "2026-03-17T00:00:01Z",
          },
          {
            id: "image-1",
            url: "https://images.example/1.jpg",
            original_filename: "front.jpg",
            is_primary: false,
            sort_order: 1,
            created_at: "2026-03-17T00:00:00Z",
          },
          {
            id: "image-3",
            url: "https://images.example/3.jpg",
            original_filename: "garden.png",
            is_primary: false,
            sort_order: 2,
            created_at: "2026-03-17T00:00:02Z",
          },
        ],
      }),
    );
    deleteListingImageMock.mockResolvedValue(
      buildListing({
        images: [
          {
            id: "image-2",
            url: "https://images.example/2.jpg",
            original_filename: "pool.jpg",
            is_primary: true,
            sort_order: 0,
            created_at: "2026-03-17T00:00:01Z",
          },
          {
            id: "image-1",
            url: "https://images.example/1.jpg",
            original_filename: "front.jpg",
            is_primary: false,
            sort_order: 1,
            created_at: "2026-03-17T00:00:00Z",
          },
        ],
      }),
    );

    renderWithProviders(<ListingForm initialListing={initialListing} listingId="listing-7" mode="edit" />);

    await screen.findByText(/front\.jpg/i);
    const fileInput = screen.getByLabelText(/choose listing image/i);
    const file = new File(["image-binary"], "garden.png", { type: "image/png" });

    fireEvent.change(fileInput, { target: { files: [file] } });
    fireEvent.click(screen.getByRole("button", { name: /upload image/i }));

    await waitFor(() => expect(uploadListingImageMock).toHaveBeenCalledWith("listing-7", file));
    expect(await screen.findByText(/garden\.png/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /move pool\.jpg earlier/i }));

    await waitFor(() => {
      expect(reorderListingImagesMock).toHaveBeenCalledWith("listing-7", ["image-2", "image-1", "image-3"]);
    });

    fireEvent.click(screen.getAllByRole("button", { name: /set primary/i })[0]);

    await waitFor(() => expect(setPrimaryListingImageMock).toHaveBeenCalledWith("listing-7", "image-2"));
    expect(await screen.findByTestId("listing-image-card-image-2")).toHaveTextContent(/primary/i);
    expect(screen.getByTestId("listing-image-card-image-1")).not.toHaveTextContent(/^primary$/i);

    fireEvent.click(within(screen.getByTestId("listing-image-card-image-3")).getByRole("button", { name: /delete image/i }));

    await waitFor(() => expect(deleteListingImageMock).toHaveBeenCalledWith("listing-7", "image-3"));
    expect(screen.queryByText(/garden\.png/i)).not.toBeInTheDocument();
  });

  it("shows backend image upload errors clearly", async () => {
    uploadListingImageMock.mockRejectedValue(
      new ApiError("invalid image file", {
        status: 400,
        traceId: "trace-image-400",
      }),
    );

    renderWithProviders(<ListingForm initialListing={buildListing()} listingId="listing-7" mode="edit" />);

    await screen.findByText(/no images yet/i);
    const fileInput = screen.getByLabelText(/choose listing image/i);
    const file = new File(["not-an-image"], "bad.txt", { type: "text/plain" });

    fireEvent.change(fileInput, { target: { files: [file] } });
    fireEvent.click(screen.getByRole("button", { name: /upload image/i }));

    await waitFor(() => expect(uploadListingImageMock).toHaveBeenCalledWith("listing-7", file));
    expect(await screen.findByRole("alert")).toHaveTextContent(/invalid image file \(trace trace-image-400\)/i);
  });

  it("renders a bootstrap failure state when categories cannot be loaded", async () => {
    getListingCategoriesMock.mockRejectedValue(new Error("categories unavailable"));

    renderWithProviders(<ListingForm mode="create" />);

    expect(await screen.findByRole("heading", { level: 2, name: /we could not prepare this listing form/i })).toBeInTheDocument();
    expect(screen.getByText(/categories unavailable/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument();
  });
});
