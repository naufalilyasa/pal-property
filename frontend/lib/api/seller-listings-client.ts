import type { ListingStatus } from "@/lib/api/listing-form";
import type { GetSellerListingsOptions, SellerListing } from "@/lib/api/seller-listings";
import { browserFetch } from "@/lib/api/browser-fetch";

export async function updateSellerListingStatus(
  listingId: string,
  status: ListingStatus,
  options: GetSellerListingsOptions = {},
): Promise<SellerListing> {
  const response = await browserFetch<SellerListing>(`/api/listings/${listingId}`, {
    method: "PUT",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ status }),
  });

  return response.data;
}

export async function deleteSellerListing(
  listingId: string,
  options: GetSellerListingsOptions = {},
): Promise<void> {
  await browserFetch<{ message: string }>(`/api/listings/${listingId}`, {
    method: "DELETE",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
  });
}
