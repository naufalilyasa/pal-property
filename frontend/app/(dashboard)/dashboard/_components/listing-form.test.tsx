import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/envelope";

import { ListingForm } from "./listing-form";

const { pushMock, refreshMock } = vi.hoisted(() => ({
  pushMock: vi.fn(),
  refreshMock: vi.fn(),
}));

const { validateListingVideoSelectionMock } = vi.hoisted(() => ({
  validateListingVideoSelectionMock: vi.fn(),
}));

const { inspectListingImageSelectionMock } = vi.hoisted(() => ({
  inspectListingImageSelectionMock: vi.fn(),
}));

const {
  createSellerListingMock,
  deleteListingImageMock,
  deleteListingVideoMock,
  getListingByIdMock,
  getListingCategoriesMock,
  getRegionCitiesMock,
  getRegionDistrictsMock,
  getRegionProvincesMock,
  getRegionVillagesMock,
  reorderListingImagesMock,
  setPrimaryListingImageMock,
  uploadListingImagesMock,
  uploadListingVideoMock,
  updateSellerListingMock,
} = vi.hoisted(() => ({
  createSellerListingMock: vi.fn(),
  deleteListingImageMock: vi.fn(),
  deleteListingVideoMock: vi.fn(),
  getListingByIdMock: vi.fn(),
  getListingCategoriesMock: vi.fn(),
  getRegionCitiesMock: vi.fn(),
  getRegionDistrictsMock: vi.fn(),
  getRegionProvincesMock: vi.fn(),
  getRegionVillagesMock: vi.fn(),
  reorderListingImagesMock: vi.fn(),
  setPrimaryListingImageMock: vi.fn(),
  uploadListingImagesMock: vi.fn(),
  uploadListingVideoMock: vi.fn(),
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
    deleteListingVideo: deleteListingVideoMock,
    getListingById: getListingByIdMock,
    getListingCategories: getListingCategoriesMock,
    getRegionCities: getRegionCitiesMock,
    getRegionDistricts: getRegionDistrictsMock,
    getRegionProvinces: getRegionProvincesMock,
    getRegionVillages: getRegionVillagesMock,
    reorderListingImages: reorderListingImagesMock,
    setPrimaryListingImage: setPrimaryListingImageMock,
    uploadListingImages: uploadListingImagesMock,
    uploadListingVideo: uploadListingVideoMock,
    updateSellerListing: updateSellerListingMock,
  };
});

vi.mock("@/features/listings/forms/listing-media", async () => {
  const actual = await vi.importActual<typeof import("@/features/listings/forms/listing-media")>(
    "@/features/listings/forms/listing-media",
  );

  return {
    ...actual,
    inspectListingImageSelection: inspectListingImageSelectionMock,
    validateListingVideoSelection: validateListingVideoSelectionMock,
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
    transaction_type: "sale",
    price: 3150000000,
    currency: "IDR",
    is_negotiable: false,
    special_offers: [],
    location_province: "Jawa Barat",
    location_province_code: "32",
    location_city: "Bandung",
    location_city_code: "32.73",
    location_district: "Cidadap",
    location_district_code: "32.73.08",
    location_village: "Hegarmanah",
    location_village_code: "32.73.08.1003",
    address_detail: "Jl. Setiabudi 10",
    latitude: null,
    longitude: null,
    bedroom_count: 5,
    bathroom_count: 3,
    floor_count: 2,
    carport_capacity: 1,
    land_area_sqm: 240,
    building_area_sqm: 180,
    certificate_type: "SHM",
    condition: "second",
    furnishing: "semi",
    electrical_power_va: 3500,
    facing_direction: "east",
    year_built: 2018,
    facilities: ["AC", "CCTV"],
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
    video: null,
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
    getRegionProvincesMock.mockResolvedValue([
      { code: "31", name: "DKI Jakarta" },
      { code: "32", name: "Jawa Barat" },
      { code: "35", name: "Jawa Timur" },
    ]);
    getRegionCitiesMock.mockImplementation(async (provinceCode: string) => {
      if (provinceCode === "31") {
        return [{ code: "31.74", name: "Jakarta Selatan" }];
      }

      if (provinceCode === "35") {
        return [{ code: "35.78", name: "Surabaya" }];
      }

      return [{ code: "32.73", name: "Bandung" }];
    });
    getRegionDistrictsMock.mockImplementation(async (cityCode: string) => {
      if (cityCode === "31.74") {
        return [{ code: "31.74.05", name: "Kebayoran Baru" }];
      }

      if (cityCode === "35.78") {
        return [{ code: "35.78.10", name: "Wonokromo" }];
      }

      return [{ code: "32.73.08", name: "Cidadap" }];
    });
    getRegionVillagesMock.mockImplementation(async (districtCode: string) => {
      if (districtCode === "31.74.05") {
        return [{ code: "31.74.05.1001", name: "Gandaria Utara" }];
      }

      if (districtCode === "35.78.10") {
        return [{ code: "35.78.10.1001", name: "Darmo" }];
      }

      return [{ code: "32.73.08.1003", name: "Hegarmanah" }];
    });
    inspectListingImageSelectionMock.mockResolvedValue({
      message: "Ready: 2 images selected (garden.png, patio.png). Recommended ratio: 4:3.",
      offRatioCount: 0,
    });
    validateListingVideoSelectionMock.mockResolvedValue({
      ok: true,
      durationSeconds: 42,
      message: "Ready: tour.mp4 · 8 MB · 42s. Backend validation still decides the final result.",
    });
  });

  it("loads create mode categories and submits the canonical create payload", async () => {
    createSellerListingMock.mockResolvedValue({ id: "listing-99" });

    renderWithProviders(<ListingForm mode="create" />);

    expect(await screen.findByRole("heading", { level: 2, name: /publish a new property draft/i })).toBeInTheDocument();
    expect(await screen.findByRole("option", { name: /house \/ villa/i })).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/^title/i), { target: { value: "Garden Residence" } });
    fireEvent.change(screen.getByLabelText(/^price/i), { target: { value: "2750000000" } });
    fireEvent.change(screen.getByLabelText(/^category/i), { target: { value: "cat-child" } });
    fireEvent.change(screen.getByLabelText(/^transaction type/i), { target: { value: "rent" } });
    fireEvent.change(await screen.findByLabelText(/^province/i), { target: { value: "31" } });
    fireEvent.change(await screen.findByLabelText(/^city/i), { target: { value: "31.74" } });
    fireEvent.change(await screen.findByLabelText(/^district/i), { target: { value: "31.74.05" } });
    fireEvent.change(await screen.findByLabelText(/village/i), { target: { value: "31.74.05.1001" } });
    fireEvent.change(screen.getByLabelText(/^bedrooms/i), { target: { value: "4" } });

    fireEvent.click(screen.getByRole("button", { name: /create listing/i }));

    await waitFor(() => {
      expect(createSellerListingMock).toHaveBeenCalledWith(
        expect.objectContaining({
          category_id: "cat-child",
          title: "Garden Residence",
          description: null,
          transaction_type: "rent",
          price: 2750000000,
          currency: "IDR",
          is_negotiable: false,
          special_offers: [],
          location_province: "DKI Jakarta",
          location_province_code: "31",
          location_city: "Jakarta Selatan",
          location_city_code: "31.74",
          location_district: "Kebayoran Baru",
          location_district_code: "31.74.05",
          location_village: "Gandaria Utara",
          location_village_code: "31.74.05.1001",
          address_detail: null,
          latitude: null,
          longitude: null,
          bedroom_count: 4,
          bathroom_count: 0,
          floor_count: null,
          carport_capacity: null,
          land_area_sqm: 0,
          building_area_sqm: 0,
          certificate_type: null,
          condition: null,
          furnishing: null,
          electrical_power_va: null,
          facing_direction: null,
          year_built: null,
          facilities: [],
          status: "active",
          specifications: {
            bedrooms: 4,
            bathrooms: 0,
            land_area_sqm: 0,
            building_area_sqm: 0,
          },
        }),
      );
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
    fireEvent.change(screen.getByLabelText(/^transaction type/i), { target: { value: "sale" } });
    fireEvent.change(await screen.findByLabelText(/^province/i), { target: { value: "35" } });
    fireEvent.change(await screen.findByLabelText(/^city/i), { target: { value: "35.78" } });
    fireEvent.change(await screen.findByLabelText(/^district/i), { target: { value: "35.78.10" } });
    fireEvent.change(await screen.findByLabelText(/village/i), { target: { value: "35.78.10.1001" } });
    fireEvent.change(screen.getByLabelText(/^address detail/i), { target: { value: "  Tower A  " } });
    fireEvent.change(screen.getByLabelText(/^status/i), { target: { value: "sold" } });
    fireEvent.change(screen.getByLabelText(/^bedrooms/i), { target: { value: "" } });
    fireEvent.change(screen.getByLabelText(/^bathrooms/i), { target: { value: "" } });
    fireEvent.change(screen.getByLabelText(/^land area/i), { target: { value: "120" } });

    fireEvent.click(screen.getByRole("button", { name: /create listing/i }));

    await waitFor(() => {
      expect(createSellerListingMock).toHaveBeenCalledWith(
        expect.objectContaining({
          category_id: null,
          title: "Sunset Loft",
          description: null,
          transaction_type: "sale",
          price: 1,
          currency: "IDR",
          is_negotiable: false,
          special_offers: [],
          location_province: "Jawa Timur",
          location_province_code: "35",
          location_city: "Surabaya",
          location_city_code: "35.78",
          location_district: "Wonokromo",
          location_district_code: "35.78.10",
          location_village: "Darmo",
          location_village_code: "35.78.10.1001",
          address_detail: "Tower A",
          latitude: null,
          longitude: null,
          bedroom_count: null,
          bathroom_count: null,
          floor_count: null,
          carport_capacity: null,
          land_area_sqm: 120,
          building_area_sqm: 0,
          certificate_type: null,
          condition: null,
          furnishing: null,
          electrical_power_va: null,
          facing_direction: null,
          year_built: null,
          facilities: [],
          status: "sold",
          specifications: {
            bedrooms: 0,
            bathrooms: 0,
            land_area_sqm: 120,
            building_area_sqm: 0,
          },
        }),
      );
    });
  });

  it("keeps media actions gated until the listing has been created", async () => {
    renderWithProviders(<ListingForm mode="create" />);

    await screen.findByRole("heading", { level: 2, name: /publish a new property draft/i });

    expect(screen.getByText(/publish the listing first, then return here to upload image batches, manage the optional video slot/i)).toBeInTheDocument();
    expect(screen.queryByTestId("listing-image-upload")).not.toBeInTheDocument();
    expect(screen.queryByTestId("listing-video-upload")).not.toBeInTheDocument();
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
    expect(screen.getByDisplayValue("Jawa Barat")).toBeInTheDocument();
    expect(screen.getByDisplayValue("5")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/^title/i), { target: { value: "Existing Residence Updated" } });
    fireEvent.change(screen.getByLabelText(/^price/i), { target: { value: "3300000000" } });
    fireEvent.change(screen.getByLabelText(/^category/i), { target: { value: "cat-root" } });
    fireEvent.change(screen.getByLabelText(/^transaction type/i), { target: { value: "rent" } });

    fireEvent.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      expect(updateSellerListingMock).toHaveBeenCalledWith(
        "listing-7",
        expect.objectContaining({
          category_id: "cat-root",
          title: "Existing Residence Updated",
          description: "Fresh paint and pool.",
          transaction_type: "rent",
          price: 3300000000,
          currency: "IDR",
          location_province: "Jawa Barat",
          location_city: "Bandung",
          location_district: "Cidadap",
          address_detail: "Jl. Setiabudi 10",
          bedroom_count: 5,
          bathroom_count: 3,
          floor_count: 2,
          carport_capacity: 1,
          land_area_sqm: 240,
          building_area_sqm: 180,
          certificate_type: "SHM",
          condition: "second",
          furnishing: "semi",
          electrical_power_va: 3500,
          facing_direction: "east",
          year_built: 2018,
          facilities: ["AC", "CCTV"],
          status: "inactive",
          specifications: {
            bedrooms: 5,
            bathrooms: 3,
            land_area_sqm: 240,
            building_area_sqm: 180,
          },
        }),
      );
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
    fireEvent.change(await screen.findByLabelText(/^province/i), { target: { value: "31" } });
    fireEvent.change(await screen.findByLabelText(/^city/i), { target: { value: "31.74" } });
    fireEvent.change(await screen.findByLabelText(/^district/i), { target: { value: "31.74.05" } });
    fireEvent.change(await screen.findByLabelText(/village/i), { target: { value: "31.74.05.1001" } });

    fireEvent.click(screen.getByRole("button", { name: /create listing/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      /title must be at least 5 characters \(trace trace-422\)/i,
    );
  });

  it("uploads multiple images, reorders, sets primary, and deletes images from backend responses", async () => {
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
    uploadListingImagesMock.mockResolvedValue(
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
    const fileInput = screen.getByLabelText(/choose listing images/i);
    const gardenFile = new File(["image-binary"], "garden.png", { type: "image/png" });
    const patioFile = new File(["image-binary"], "patio.png", { type: "image/png" });

    fireEvent.change(fileInput, { target: { files: [gardenFile, patioFile] } });
    expect(screen.getByText(/ready: 2 images selected/i)).toBeInTheDocument();
    expect(await screen.findByText(/recommended ratio: 4:3/i)).toBeInTheDocument();

    const uploadImagesButton = await screen.findByRole("button", { name: /upload images/i });
    await waitFor(() => expect(uploadImagesButton).toBeEnabled());
    fireEvent.click(uploadImagesButton);

    await waitFor(() => expect(uploadListingImagesMock).toHaveBeenCalledWith("listing-7", [gardenFile, patioFile]));
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
    uploadListingImagesMock.mockRejectedValue(
      new ApiError("invalid image file", {
        status: 400,
        traceId: "trace-image-400",
      }),
    );

    renderWithProviders(<ListingForm initialListing={buildListing()} listingId="listing-7" mode="edit" />);

    await screen.findByText(/no images yet/i);
    const fileInput = screen.getByLabelText(/choose listing images/i);
    const file = new File(["not-an-image"], "bad.txt", { type: "text/plain" });
    inspectListingImageSelectionMock.mockResolvedValueOnce({
      message: "Ready: bad.txt. Recommended ratio: 4:3.",
      offRatioCount: 0,
    });

    fireEvent.change(fileInput, { target: { files: [file] } });

    const uploadImagesButton = await screen.findByRole("button", { name: /upload images/i });
    await waitFor(() => expect(uploadImagesButton).toBeEnabled());
    fireEvent.click(uploadImagesButton);

    await waitFor(() => expect(uploadListingImagesMock).toHaveBeenCalledWith("listing-7", [file]));
    expect(await screen.findByRole("alert")).toHaveTextContent(/invalid image file \(trace trace-image-400\)/i);
  });

  it("shows a client-side video validation message before upload", async () => {
    validateListingVideoSelectionMock.mockResolvedValue({
      ok: false,
      durationSeconds: null,
      message: "Choose a video under 100 MB before uploading. Backend validation still decides the final result.",
    });

    renderWithProviders(<ListingForm initialListing={buildListing()} listingId="listing-7" mode="edit" />);

    await screen.findByText(/optional listing video/i);
    const file = new File(["video-stream"], "tour.mp4", { type: "video/mp4" });

    fireEvent.change(screen.getByTestId("listing-video-upload"), { target: { files: [file] } });

    await waitFor(() => expect(validateListingVideoSelectionMock).toHaveBeenCalledWith(file));
    expect(await screen.findByTestId("listing-video-error")).toHaveTextContent(/under 100 mb/i);
    expect(uploadListingVideoMock).not.toHaveBeenCalled();
  });

  it("uploads and deletes a single listing video while keeping image state intact", async () => {
    validateListingVideoSelectionMock.mockResolvedValue({
      ok: true,
      durationSeconds: 42,
      message: "Ready: walkthrough.mp4 · 8 MB · 42s. Backend validation still decides the final result.",
    });

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
      ],
    });

    uploadListingVideoMock.mockResolvedValue(
      buildListing({
        images: initialListing.images,
        video: {
          id: "video-1",
          url: "https://videos.example/tour.mp4",
          original_filename: "walkthrough.mp4",
          duration_seconds: 42,
          created_at: "2026-03-17T00:00:04Z",
        },
      }),
    );
    deleteListingVideoMock.mockResolvedValue(
      buildListing({
        images: initialListing.images,
        video: null,
      }),
    );

    renderWithProviders(<ListingForm initialListing={initialListing} listingId="listing-7" mode="edit" />);

    await screen.findByText(/front\.jpg/i);
    const video = new File(["video-stream"], "walkthrough.mp4", { type: "video/mp4" });

    fireEvent.change(screen.getByTestId("listing-video-upload"), { target: { files: [video] } });

    await waitFor(() => expect(validateListingVideoSelectionMock).toHaveBeenCalledWith(video));
    expect(await screen.findByText(/ready: walkthrough\.mp4/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /upload video/i }));

    await waitFor(() => expect(uploadListingVideoMock).toHaveBeenCalledWith("listing-7", video));
    expect(await screen.findByText(/walkthrough\.mp4/i)).toBeInTheDocument();
    expect(screen.getByText(/front\.jpg/i)).toBeInTheDocument();
    expect(screen.getByText(/delete this video before selecting a replacement/i)).toBeInTheDocument();
    expect(screen.getByTestId("listing-video-upload")).toBeDisabled();

    fireEvent.click(screen.getByRole("button", { name: /delete video/i }));

    await waitFor(() => expect(deleteListingVideoMock).toHaveBeenCalledWith("listing-7"));
    expect(await screen.findByText(/no video yet/i)).toBeInTheDocument();
    expect(screen.queryByText(/walkthrough\.mp4/i)).not.toBeInTheDocument();
    expect(screen.getByText(/front\.jpg/i)).toBeInTheDocument();
  });

  it("renders a bootstrap failure state when categories cannot be loaded", async () => {
    getListingCategoriesMock.mockRejectedValue(new Error("categories unavailable"));

    renderWithProviders(<ListingForm mode="create" />);

    expect(await screen.findByRole("heading", { level: 2, name: /we could not prepare this listing form/i })).toBeInTheDocument();
    expect(screen.getByText(/categories unavailable/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument();
  });
});
