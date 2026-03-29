import { describe, expect, it, vi, beforeEach, afterAll } from "vitest";

import { ApiError } from "@/lib/api/envelope";
import * as browserFetchModule from "@/lib/api/browser-fetch";
import * as serverFetchModule from "@/lib/api/server-fetch";

import { getSavedListingIds, getSavedListings, saveListing } from "@/lib/api/saved-listings";

const browserFetchSpy = vi.spyOn(browserFetchModule, "browserFetch");
const serverFetchSpy = vi.spyOn(serverFetchModule, "serverFetch");

describe("saved listings api", () => {
  beforeEach(() => {
    browserFetchSpy.mockReset();
    serverFetchSpy.mockReset();
  });

  afterAll(() => {
    browserFetchSpy.mockRestore();
    serverFetchSpy.mockRestore();
  });

  it("normalizes toggle responses when saving a listing", async () => {
    browserFetchSpy.mockResolvedValueOnce({
      data: { listing_id: "listing-1", saved: true },
      message: "ok",
      traceId: "trace-1",
    });

    const result = await saveListing("listing-1", { baseUrl: "https://api" });

    expect(result).toEqual({ listingId: "listing-1", saved: true });
    expect(browserFetchSpy).toHaveBeenCalledWith(
      "/api/me/saved-listings",
      expect.objectContaining({
        method: "POST",
        cache: "no-store",
        baseUrl: "https://api",
        headers: expect.objectContaining({
          "Content-Type": "application/json",
        }),
        body: JSON.stringify({ listing_id: "listing-1" }),
      }),
    );
  });

  it("surfaces ApiError failures when saving listings", async () => {
    const apiError = new ApiError("failed", { status: 400, traceId: "trace-error" });
    browserFetchSpy.mockRejectedValueOnce(apiError);

    await expect(saveListing("listing-1")).rejects.toBe(apiError);
  });

  it("fetches saved listing ids with serverFetch when a cookie header is supplied", async () => {
    serverFetchSpy.mockResolvedValueOnce({
      data: { listing_ids: ["listing-1"] },
      message: "ok",
      traceId: "trace-ids",
    });

    const result = await getSavedListingIds(["listing-1"], { cookieHeader: "cookie" });

    expect(result).toEqual({ listingIds: ["listing-1"] });
    expect(serverFetchSpy).toHaveBeenCalledWith(
      expect.stringContaining("/api/me/saved-listings/contains"),
      expect.objectContaining({
        method: "GET",
        cache: "no-store",
        cookieHeader: "cookie",
      }),
    );
  });

  it("queries saved listings via serverFetch when pagination params are present", async () => {
    serverFetchSpy.mockResolvedValueOnce({
      data: { data: [], total: 0, page: 2, limit: 5, total_pages: 0 },
      message: "ok",
      traceId: "trace-listings",
    });

    const response = await getSavedListings({ page: "2", limit: "5" }, { cookieHeader: "cookie" });

    expect(response).toEqual({
      data: [],
      total: 0,
      page: 2,
      limit: 5,
      total_pages: 0,
    });
    expect(serverFetchSpy).toHaveBeenCalledWith(
      "/api/me/saved-listings?page=2&limit=5",
      expect.objectContaining({
        method: "GET",
        cache: "no-store",
        cookieHeader: "cookie",
      }),
    );
  });
});
