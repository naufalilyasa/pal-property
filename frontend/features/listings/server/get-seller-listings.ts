import "server-only";

import { serverFetch } from "@/lib/api/server-fetch";
import type { SellerListing, SellerListingsPage } from "@/lib/api/seller-listings";
import { ApiError } from "@/lib/api/envelope";
import { getRequestCookieHeader } from "@/lib/server/cookies";

export async function getSellerListingsPage(limit = 1000) {
  const response = await serverFetch<SellerListingsPage>(`/auth/me/listings?limit=${limit}`, {
    method: "GET",
    cache: "no-store",
    cookieHeader: await getRequestCookieHeader(),
  });

  return response.data;
}

export async function getSellerListingById(listingId: string): Promise<SellerListing> {
  const page = await getSellerListingsPage();
  const listing = page.data.find((candidate) => candidate.id === listingId);

  if (!listing) {
    throw new ApiError("Listing not found.", {
      status: 404,
    });
  }

  return listing;
}
