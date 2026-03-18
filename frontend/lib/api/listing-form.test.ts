import { describe, expect, it, vi } from "vitest";

import {
  formatListingFormError,
  getDefaultSpecifications,
  getListingCategories,
  parseListingSpecifications,
  deleteListingImage,
  reorderListingImages,
  setPrimaryListingImage,
  uploadListingImage,
} from "@/lib/api/listing-form";
import { ApiError } from "@/lib/api/envelope";

const backendEnvelope = <T,>(data: T, message = "Success") => ({
  success: true,
  message,
  data,
  trace_id: "trace-listing-form",
});

describe("listing image api helpers", () => {
  it("flattens nested categories into root and parent/child options", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify(
          backendEnvelope([
            {
              id: "cat-root",
              name: "House",
              slug: "house",
              parent_id: null,
              icon_url: null,
              children: [
                {
                  id: "cat-child",
                  name: "Villa",
                  slug: "villa",
                  icon_url: null,
                },
              ],
            },
            {
              id: "cat-apartment",
              name: "Apartment",
              slug: "apartment",
              parent_id: null,
              icon_url: null,
              children: [],
            },
          ]),
        ),
        {
          status: 200,
          headers: { "Content-Type": "application/json" },
        },
      ),
    );

    await expect(
      getListingCategories({
        baseUrl: "http://127.0.0.1:8080",
        fetch: fetchMock,
      }),
    ).resolves.toEqual([
      { id: "cat-root", name: "House", slug: "house", label: "House" },
      { id: "cat-child", name: "Villa", slug: "villa", label: "House / Villa" },
      { id: "cat-apartment", name: "Apartment", slug: "apartment", label: "Apartment" },
    ]);
  });

  it("normalizes malformed listing specifications and formats errors with traces", () => {
    expect(getDefaultSpecifications()).toEqual({
      bedrooms: 0,
      bathrooms: 0,
      land_area_sqm: 0,
      building_area_sqm: 0,
    });

    expect(
      parseListingSpecifications({
        bedrooms: 3.9,
        bathrooms: -2,
        land_area_sqm: Number.POSITIVE_INFINITY,
        building_area_sqm: 180,
      }),
    ).toEqual({
      bedrooms: 3,
      bathrooms: 0,
      land_area_sqm: 0,
      building_area_sqm: 180,
    });

    expect(parseListingSpecifications(null)).toEqual(getDefaultSpecifications());

    expect(
      formatListingFormError(
        new ApiError("invalid category", {
          status: 400,
          traceId: "trace-category-400",
        }),
      ),
    ).toBe("invalid category (trace trace-category-400)");
    expect(formatListingFormError(new Error("network timeout"))).toBe("network timeout");
    expect(formatListingFormError(42)).toBe("We could not complete the listing request.");
  });

  it("uploads an image with multipart form data", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify(backendEnvelope({ id: "listing-1", images: [] })), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const file = new File(["image-binary"], "front-yard.png", { type: "image/png" });

    await uploadListingImage("listing-1", file, {
      baseUrl: "http://127.0.0.1:8080",
      fetch: fetchMock,
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://127.0.0.1:8080/api/listings/listing-1/images",
      expect.objectContaining({
        method: "POST",
        credentials: "include",
        body: expect.any(FormData),
      }),
    );

    const [, init] = fetchMock.mock.calls[0];
    expect(init?.body).toBeInstanceOf(FormData);
    expect((init?.body as FormData).get("file")).toBe(file);
  });

  it("targets delete, primary, and reorder image routes with the backend contract", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockImplementation(async () =>
      new Response(JSON.stringify(backendEnvelope({ id: "listing-1", images: [] })), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    await deleteListingImage("listing-1", "image-2", {
      baseUrl: "http://127.0.0.1:8080",
      fetch: fetchMock,
    });
    await setPrimaryListingImage("listing-1", "image-3", {
      baseUrl: "http://127.0.0.1:8080",
      fetch: fetchMock,
    });
    await reorderListingImages("listing-1", ["image-3", "image-1", "image-2"], {
      baseUrl: "http://127.0.0.1:8080",
      fetch: fetchMock,
    });

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "http://127.0.0.1:8080/api/listings/listing-1/images/image-2",
      expect.objectContaining({ method: "DELETE" }),
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "http://127.0.0.1:8080/api/listings/listing-1/images/image-3/primary",
      expect.objectContaining({ method: "PATCH" }),
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "http://127.0.0.1:8080/api/listings/listing-1/images/reorder",
      expect.objectContaining({
        method: "PATCH",
        body: JSON.stringify({ ordered_image_ids: ["image-3", "image-1", "image-2"] }),
      }),
    );
  });
});
